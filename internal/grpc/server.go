package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	nexusconfig "nexus/internal/config"
	nexusjwt "nexus/internal/pkg/jwt"
	nodepb "nexus/internal/proto/node"
	nexussvc "nexus/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ---------------------------------------------------------------------------
// Context key for the authenticated node ID
// ---------------------------------------------------------------------------

type ctxKey struct{}

var nodeIDKey ctxKey

func nodeIDFromCtx(ctx context.Context) (uint, bool) {
	v, ok := ctx.Value(nodeIDKey).(uint)
	return v, ok
}

// ---------------------------------------------------------------------------
// In-memory config-sync tracker
// ---------------------------------------------------------------------------

// lastSyncMap records the UpdatedAt timestamp the node last saw, so the
// heartbeat handler can detect admin-side config changes.
var lastSyncMap sync.Map // map[uint]time.Time

// ---------------------------------------------------------------------------
// Server struct
// ---------------------------------------------------------------------------

type grpcServer struct {
	nodepb.UnimplementedNodeServiceServer
}

// ---------------------------------------------------------------------------
// StartGRPCServer boots the gRPC listener and blocks until the server stops.
// It optionally enables TLS when cert_file and key_file are set in config.
// ---------------------------------------------------------------------------

func StartGRPCServer(cfg nexusconfig.GRPCConfig) error {
	lis, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		return fmt.Errorf("gRPC listen on %s: %w", cfg.Listen, err)
	}

	var opts []grpc.ServerOption

	// Unary interceptor - validates auth_token for all RPCs except Register.
	opts = append(opts, grpc.UnaryInterceptor(authInterceptor()))

	// Optional TLS.
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return fmt.Errorf("load TLS cert: %w", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	s := grpc.NewServer(opts...)
	nodepb.RegisterNodeServiceServer(s, &grpcServer{})

	log.Printf("gRPC node service listening on %s", cfg.Listen)
	return s.Serve(lis)
}

// ---------------------------------------------------------------------------
// Auth interceptor
// ---------------------------------------------------------------------------

func authInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// Register is the only RPC that does not require authentication.
		if info.FullMethod == "/nexus.node.NodeService/Register" {
			return handler(ctx, req)
		}

		token, err := extractToken(ctx)
		if err != nil {
			return nil, err
		}

		claims, err := nexusjwt.Parse(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid auth token")
		}

		// We reuse the user_id claim to carry the node DB primary key.
		ctx = context.WithValue(ctx, nodeIDKey, claims.UserID)
		return handler(ctx, req)
	}
}

// extractToken pulls the Bearer token from gRPC metadata.
func extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}
	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing auth token")
	}
	token := values[0]
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}
	if token == "" {
		return "", status.Error(codes.Unauthenticated, "empty auth token")
	}
	return token, nil
}

// ---------------------------------------------------------------------------
// Register
// ---------------------------------------------------------------------------

func (s *grpcServer) Register(ctx context.Context, req *nodepb.RegisterRequest) (*nodepb.RegisterResponse, error) {
	if req.Token == "" {
		return &nodepb.RegisterResponse{Success: false, Error: "register token is required"}, nil
	}

	node, err := nexussvc.GetNodeByToken(req.Token)
	if err != nil {
		return &nodepb.RegisterResponse{Success: false, Error: "invalid register token"}, nil
	}

	// Update the node reported name and address.
	if err := nexussvc.UpdateNodeInfo(node.ID, req.NodeName, req.Address); err != nil {
		return &nodepb.RegisterResponse{Success: false, Error: "failed to update node info"}, nil
	}

	// Generate a JWT that carries the node ID. The standard jwt.Generate
	// puts the id in the user_id claim - we reuse that for node auth.
	expireHours := nexusconfig.Global.JWT.ExpireHours
	if expireHours == 0 {
		expireHours = 720 // 30 days default for nodes
	}
	authToken, err := nexusjwt.Generate(node.ID, false, expireHours)
	if err != nil {
		return &nodepb.RegisterResponse{Success: false, Error: "failed to generate auth token"}, nil
	}

	log.Printf("node registered: id=%d  name=%s  address=%s", node.ID, req.NodeName, req.Address)

	return &nodepb.RegisterResponse{
		Success:   true,
		NodeID:    strconv.FormatUint(uint64(node.ID), 10),
		AuthToken: authToken,
	}, nil
}

// ---------------------------------------------------------------------------
// Heartbeat
// ---------------------------------------------------------------------------

func (s *grpcServer) Heartbeat(ctx context.Context, req *nodepb.HeartbeatRequest) (*nodepb.HeartbeatResponse, error) {
	nodeID, ok := nodeIDFromCtx(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unrecognized node identity")
	}

	// Persist the latest metrics.
	if err := nexussvc.UpdateNodeStatus(nodeID, true, req.CpuUsage, req.MemoryUsage); err != nil {
		return nil, status.Errorf(codes.Internal, "update node status: %v", err)
	}

	// Determine whether the admin has changed the node config since the
	// node last fetched it.
	configChanged := false
	node, err := nexussvc.GetNodeByID(nodeID)
	if err == nil {
		if lastRaw, loaded := lastSyncMap.Load(nodeID); loaded {
			lastSync := lastRaw.(time.Time)
			if node.UpdatedAt.After(lastSync) {
				configChanged = true
			}
		} else {
			// First heartbeat - tell the node to fetch config.
			configChanged = true
		}
	}

	return &nodepb.HeartbeatResponse{
		Success:       true,
		ConfigChanged: configChanged,
	}, nil
}

// ---------------------------------------------------------------------------
// GetConfig
// ---------------------------------------------------------------------------

func (s *grpcServer) GetConfig(ctx context.Context, req *nodepb.GetConfigRequest) (*nodepb.GetConfigResponse, error) {
	nodeID, ok := nodeIDFromCtx(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unrecognized node identity")
	}

	node, err := nexussvc.GetNodeByID(nodeID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "node not found: %v", err)
	}

	users, err := nexussvc.GetActiveUsers()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fetch active users: %v", err)
	}

	singboxCfg, err := nexussvc.GenerateSingboxConfig(*node, users)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate sing-box config: %v", err)
	}

	usersJSON, err := json.Marshal(users)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "marshal users: %v", err)
	}

	// Record the sync timestamp so Heartbeat can detect subsequent changes.
	lastSyncMap.Store(nodeID, time.Now())

	return &nodepb.GetConfigResponse{
		SingboxConfig: singboxCfg,
		UsersJSON:     string(usersJSON),
	}, nil
}

// ---------------------------------------------------------------------------
// ReportTraffic
// ---------------------------------------------------------------------------

func (s *grpcServer) ReportTraffic(ctx context.Context, req *nodepb.TrafficReport) (*nodepb.TrafficReportResponse, error) {
	nodeID, ok := nodeIDFromCtx(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unrecognized node identity")
	}

	entries := make([]nexussvc.TrafficEntry, len(req.Entries))
	for i, e := range req.Entries {
		entries[i] = nexussvc.TrafficEntry{
			UserUUID: e.UserUUID,
			Upload:   e.Upload,
			Download: e.Download,
		}
	}

	if err := nexussvc.RecordTraffic(nodeID, entries); err != nil {
		return nil, status.Errorf(codes.Internal, "record traffic: %v", err)
	}

	return &nodepb.TrafficReportResponse{Success: true}, nil
}
