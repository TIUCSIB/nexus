<script setup lang="ts">
import { ref } from 'vue'
import { Menu, X, Search, Play, Plus, Flame } from 'lucide-vue-next'

// 门户品牌名（占位，可按需替换）
const siteName = '次元动漫'

const mobileOpen = ref(false)

// 轻量本地提示：点击伪装按钮时给出"建设中"反馈，不做任何跳转
const toastMsg = ref('')
let toastTimer: ReturnType<typeof setTimeout> | undefined
function soon(msg = '该模块正在建设中，敬请期待～') {
  toastMsg.value = msg
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => (toastMsg.value = ''), 2200)
}

// 图片加载失败 → 隐藏 img，露出底层渐变兜底
function onImgError(e: Event) {
  ;(e.target as HTMLImageElement).style.display = 'none'
}

interface Anime {
  title: string
  eps: string
  gradient: string
  tag?: string
  image?: string
}
interface RankItem {
  rank: number
  title: string
  heat: string
}
interface ScheduleItem {
  day: string
  title: string
  time: string
}

const navLinks = ['首页', '番剧', '国创', '资讯', '社区', '排行榜']

// featured.image 换成真实番剧封面 URL 即可（当前为演示用稳定实图）
const featured = {
  title: '苍穹之翼',
  tags: ['热血', '科幻', '冒险'],
  desc: '在崩坏的苍穹之下，少年与机械之翼缔结契约，踏上找回失落天空的征途。每周四准时更新，燃向王炸开场。',
  eps: '更新至 14 话',
  score: '9.2',
  gradient: 'linear-gradient(135deg,#fb7299 0%,#7b5cff 100%)',
  image: 'https://picsum.photos/seed/nexus-anime-featured/900/1200',
}

const todayUpdates: Anime[] = [
  { title: '苍穹之翼', eps: '更新至 14 话', gradient: 'linear-gradient(135deg,#fb7299,#7b5cff)', tag: '热血', image: 'https://picsum.photos/seed/nexus-anime-1/600/800' },
  { title: '星海征途', eps: '更新至 08 话', gradient: 'linear-gradient(135deg,#4facfe,#00f2fe)', tag: '科幻', image: 'https://picsum.photos/seed/nexus-anime-2/600/800' },
  { title: '幻夜剑歌', eps: '更新至 21 话', gradient: 'linear-gradient(135deg,#667eea,#764ba2)', tag: '战斗', image: 'https://picsum.photos/seed/nexus-anime-3/600/800' },
  { title: '樱色协奏曲', eps: '更新至 06 话', gradient: 'linear-gradient(135deg,#ff9a9e,#fecfef)', tag: '恋爱', image: 'https://picsum.photos/seed/nexus-anime-4/600/800' },
  { title: '末日防线', eps: '更新至 11 话', gradient: 'linear-gradient(135deg,#43e97b,#38f9d7)', tag: '末日', image: 'https://picsum.photos/seed/nexus-anime-5/600/800' },
  { title: '妖精森林', eps: '已完结', gradient: 'linear-gradient(135deg,#84fab0,#8fd3f4)', tag: '治愈', image: 'https://picsum.photos/seed/nexus-anime-6/600/800' },
  { title: '都市妖奇谭', eps: '更新至 03 话', gradient: 'linear-gradient(135deg,#fccb90,#d57eeb)', tag: '奇幻', image: 'https://picsum.photos/seed/nexus-anime-7/600/800' },
  { title: '深海之歌', eps: '更新至 09 话', gradient: 'linear-gradient(135deg,#30cfd0,#330867)', tag: '音乐', image: 'https://picsum.photos/seed/nexus-anime-8/600/800' },
]

const ranking: RankItem[] = [
  { rank: 1, title: '苍穹之翼', heat: '328.6万' },
  { rank: 2, title: '幻夜剑歌', heat: '265.1万' },
  { rank: 3, title: '星海征途', heat: '241.8万' },
  { rank: 4, title: '都市妖奇谭', heat: '187.3万' },
  { rank: 5, title: '樱色协奏曲', heat: '160.9万' },
  { rank: 6, title: '末日防线', heat: '143.2万' },
  { rank: 7, title: '深海之歌', heat: '121.5万' },
  { rank: 8, title: '妖精森林', heat: '98.7万' },
  { rank: 9, title: '剑与花的物语', heat: '84.0万' },
  { rank: 10, title: '云端之上', heat: '72.4万' },
]

const schedule: ScheduleItem[] = [
  { day: '周一', title: '妖精森林', time: '20:00' },
  { day: '周二', title: '都市妖奇谭', time: '19:30' },
  { day: '周三', title: '深海之歌', time: '21:00' },
  { day: '周四', title: '苍穹之翼', time: '22:00' },
  { day: '周五', title: '剑与花的物语', time: '20:30' },
  { day: '周六', title: '樱色协奏曲', time: '18:00' },
  { day: '周日', title: '星海征途', time: '19:00' },
]

function rankColor(rank: number) {
  if (rank === 1) return 'text-[#ff5c8a]'
  if (rank === 2) return 'text-[#ff9a3d]'
  if (rank === 3) return 'text-[#ffce4d]'
  return 'text-[#9499a0]'
}
</script>

<template>
  <div class="min-h-svh bg-[#f6f7fb] text-[#18191c]" style="font-family: 'Inter', ui-sans-serif, system-ui, sans-serif">
    <!-- 顶部导航 -->
    <header class="sticky top-0 z-50 border-b border-[#e3e5e7] bg-white/90 backdrop-blur">
      <div class="mx-auto flex h-16 max-w-6xl items-center gap-6 px-5">
        <a href="/" class="flex shrink-0 items-center gap-2">
          <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-[#fb7299] text-white">
            <Play class="h-4 w-4" />
          </span>
          <span class="text-lg font-bold tracking-tight">{{ siteName }}</span>
        </a>

        <nav class="hidden items-center gap-1 md:flex">
          <a
            v-for="(link, i) in navLinks"
            :key="link"
            href="#"
            @click.prevent="soon()"
            class="rounded-md px-3 py-2 text-sm font-medium transition-colors"
            :class="i === 0 ? 'bg-[#fff0f5] text-[#fb7299]' : 'text-[#61666d] hover:bg-[#f6f7fb] hover:text-[#18191c]'"
          >
            {{ link }}
          </a>
        </nav>

        <div class="ml-auto flex items-center gap-3">
          <div class="hidden items-center rounded-full border border-[#e3e5e7] bg-[#f6f7fb] px-3 py-1.5 lg:flex">
            <input
              type="text"
              placeholder="搜索番剧、角色、声优…"
              class="w-44 bg-transparent text-sm text-[#18191c] outline-none placeholder:text-[#9499a0]"
            />
            <button type="button" class="text-[#9499a0] hover:text-[#fb7299]" @click="soon()" aria-label="搜索">
              <Search class="h-4 w-4" />
            </button>
          </div>
          <button
            type="button"
            class="rounded-full px-4 py-1.5 text-sm font-medium text-[#61666d] transition-colors hover:text-[#18191c]"
            @click="soon()"
          >
            登录
          </button>
          <button
            type="button"
            class="rounded-full bg-[#fb7299] px-4 py-1.5 text-sm font-medium text-white transition-colors hover:bg-[#fc8bab] active:scale-95"
            @click="soon()"
          >
            注册
          </button>
          <button
            type="button"
            class="inline-flex h-9 w-9 items-center justify-center rounded-md text-[#61666d] md:hidden"
            :aria-label="mobileOpen ? '关闭菜单' : '打开菜单'"
            @click="mobileOpen = !mobileOpen"
          >
            <component :is="mobileOpen ? X : Menu" class="h-5 w-5" />
          </button>
        </div>
      </div>

      <!-- 移动端菜单 -->
      <div v-if="mobileOpen" class="border-t border-[#e3e5e7] bg-white md:hidden">
        <div class="mx-auto flex max-w-6xl flex-col gap-1 px-5 py-3">
          <a
            v-for="(link, i) in navLinks"
            :key="link"
            href="#"
            @click.prevent="soon(); mobileOpen = false"
            class="rounded-md px-3 py-2 text-sm font-medium"
            :class="i === 0 ? 'bg-[#fff0f5] text-[#fb7299]' : 'text-[#61666d] hover:bg-[#f6f7fb]'"
          >
            {{ link }}
          </a>
        </div>
      </div>
    </header>

    <!-- 主视觉 / 今日推荐（暗色动漫风） -->
    <section class="mx-auto max-w-6xl px-5 py-10">
      <div class="relative overflow-hidden rounded-2xl border border-white/10 shadow-xl">
        <!-- 暗色夜空底 -->
        <div class="absolute inset-0" style="background: linear-gradient(135deg,#1a1033 0%,#3a1a6b 55%,#0f0524 100%)"></div>
        <!-- 霓虹光晕 -->
        <div class="absolute -left-16 -top-16 h-72 w-72 rounded-full bg-[#fb7299]/30 blur-3xl"></div>
        <div class="absolute -right-16 bottom-0 h-80 w-80 rounded-full bg-[#23ade5]/25 blur-3xl"></div>

        <div class="relative grid md:grid-cols-2">
          <!-- 左：信息 -->
          <div class="flex flex-col justify-center gap-4 p-8 md:p-10">
            <span class="inline-flex w-fit items-center gap-1 rounded-full bg-white/10 px-3 py-1 text-xs font-semibold text-[#ff9ec4]">
              <Flame class="h-3.5 w-3.5" /> 今日推荐
            </span>
            <h1
              class="text-3xl font-bold tracking-tight text-white md:text-4xl"
              style="text-shadow: 0 0 24px rgba(251,114,153,0.45)"
            >
              {{ featured.title }}
            </h1>
            <div class="flex flex-wrap gap-2">
              <span
                v-for="t in featured.tags"
                :key="t"
                class="rounded-md bg-white/10 px-2.5 py-1 text-xs font-medium text-white/80"
              >
                {{ t }}
              </span>
            </div>
            <p class="max-w-md text-sm leading-relaxed text-white/70">{{ featured.desc }}</p>
            <div class="flex items-center gap-4 text-sm text-white/60">
              <span>{{ featured.eps }}</span>
              <span class="text-[#ff9ec4]">★ {{ featured.score }}</span>
            </div>
            <div class="mt-2 flex gap-3">
              <button
                type="button"
                class="inline-flex items-center gap-1.5 rounded-full bg-[#fb7299] px-6 py-2.5 text-sm font-medium text-white shadow-[0_0_24px_rgba(251,114,153,0.45)] transition-colors hover:bg-[#fc8bab] active:scale-95"
                @click="soon()"
              >
                <Play class="h-4 w-4" /> 立即观看
              </button>
              <button
                type="button"
                class="inline-flex items-center gap-1.5 rounded-full border border-white/25 px-6 py-2.5 text-sm font-medium text-white transition-colors hover:bg-white/10"
                @click="soon()"
              >
                <Plus class="h-4 w-4" /> 加入追番
              </button>
            </div>
          </div>

          <!-- 右：海报 -->
          <div class="relative min-h-[280px] md:min-h-full" :style="{ background: featured.gradient }">
            <img
              v-if="featured.image"
              :src="featured.image"
              alt=""
              class="absolute inset-0 h-full w-full object-cover opacity-95"
              @error="onImgError"
            />
            <div class="absolute bottom-4 left-4 right-4 rounded-lg bg-black/35 px-3 py-2 text-center text-sm font-semibold text-white backdrop-blur">
              {{ featured.title }}
            </div>
            <div class="absolute right-4 top-4 rounded-full bg-black/35 px-3 py-1 text-xs font-medium text-white backdrop-blur">
              {{ featured.eps }}
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 今日更新 -->
    <section class="mx-auto max-w-6xl px-5 py-6">
      <div class="mb-5 flex items-end justify-between">
        <h2 class="text-xl font-bold tracking-tight">今日更新</h2>
        <button type="button" class="text-sm text-[#9499a0] hover:text-[#fb7299]" @click="soon()">查看更多 ›</button>
      </div>
      <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
        <div
          v-for="a in todayUpdates"
          :key="a.title"
          class="group cursor-pointer"
          @click="soon()"
        >
          <div class="relative aspect-[3/4] overflow-hidden rounded-lg" :style="{ background: a.gradient }">
            <img
              v-if="a.image"
              :src="a.image"
              alt=""
              class="absolute inset-0 h-full w-full object-cover"
              @error="onImgError"
            />
            <div class="absolute left-2 top-2 rounded bg-black/35 px-2 py-0.5 text-[11px] font-medium text-white backdrop-blur">
              {{ a.tag }}
            </div>
            <div class="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/65 to-transparent p-2.5 pt-6">
              <div class="truncate text-sm font-semibold text-white">{{ a.title }}</div>
              <div class="truncate text-[11px] text-white/80">{{ a.eps }}</div>
            </div>
          </div>
          <div class="mt-2 truncate text-sm font-medium text-[#18191c] group-hover:text-[#fb7299]">{{ a.title }}</div>
        </div>
      </div>
    </section>

    <!-- 热门排行榜 -->
    <section class="mx-auto max-w-6xl px-5 py-6">
      <div class="mb-5 flex items-end justify-between">
        <h2 class="text-xl font-bold tracking-tight">热门排行榜</h2>
        <button type="button" class="text-sm text-[#9499a0] hover:text-[#fb7299]" @click="soon()">完整榜单 ›</button>
      </div>
      <div class="overflow-hidden rounded-xl border border-[#e3e5e7] bg-white">
        <div
          v-for="r in ranking"
          :key="r.rank"
          class="flex items-center gap-4 border-b border-[#f1f2f3] px-5 py-3 last:border-0 hover:bg-[#fafbfc]"
          @click="soon()"
        >
          <span class="w-6 text-center text-lg font-bold tabular-nums" :class="rankColor(r.rank)">{{ r.rank }}</span>
          <span class="flex-1 truncate text-sm font-medium text-[#18191c]">{{ r.title }}</span>
          <span class="text-xs text-[#9499a0]">{{ r.heat }}</span>
        </div>
      </div>
    </section>

    <!-- 新番时间表 -->
    <section class="mx-auto max-w-6xl px-5 py-6">
      <h2 class="mb-5 text-xl font-bold tracking-tight">新番时间表</h2>
      <div class="grid grid-cols-2 gap-3 sm:grid-cols-4 md:grid-cols-7">
        <div
          v-for="s in schedule"
          :key="s.day"
          class="rounded-xl border border-[#e3e5e7] bg-white p-4 text-center transition-colors hover:border-[#fb7299]"
          @click="soon()"
        >
          <div class="text-sm font-semibold text-[#18191c]">{{ s.day }}</div>
          <div class="mt-2 truncate text-xs font-medium text-[#fb7299]">{{ s.title }}</div>
          <div class="mt-1 text-[11px] text-[#9499a0]">{{ s.time }}</div>
        </div>
      </div>
    </section>

    <!-- 页脚 -->
    <footer class="mt-8 border-t border-[#e3e5e7] bg-white">
      <div class="mx-auto grid max-w-6xl gap-8 px-5 py-12 md:grid-cols-4">
        <div class="md:col-span-1">
          <div class="flex items-center gap-2">
            <span class="flex h-7 w-7 items-center justify-center rounded-lg bg-[#fb7299] text-white">
              <Play class="h-3.5 w-3.5" />
            </span>
            <span class="text-base font-bold">{{ siteName }}</span>
          </div>
          <p class="mt-3 text-sm leading-relaxed text-[#9499a0]">专注于番剧、国创与动漫资讯的二次元 portal，陪你追每一部好作品。</p>
        </div>
        <div>
          <div class="text-xs font-semibold uppercase tracking-wide text-[#9499a0]">追番</div>
          <ul class="mt-4 space-y-2 text-sm text-[#61666d]">
            <li><a href="#" @click.prevent="soon()" class="hover:text-[#fb7299]">番剧库</a></li>
            <li><a href="#" @click.prevent="soon()" class="hover:text-[#fb7299]">新番时间表</a></li>
            <li><a href="#" @click.prevent="soon()" class="hover:text-[#fb7299]">排行榜</a></li>
          </ul>
        </div>
        <div>
          <div class="text-xs font-semibold uppercase tracking-wide text-[#9499a0]">社区</div>
          <ul class="mt-4 space-y-2 text-sm text-[#61666d]">
            <li><a href="#" @click.prevent="soon()" class="hover:text-[#fb7299]">资讯动态</a></li>
            <li><a href="#" @click.prevent="soon()" class="hover:text-[#fb7299]">讨论区</a></li>
            <li><a href="#" @click.prevent="soon()" class="hover:text-[#fb7299]">创作者中心</a></li>
          </ul>
        </div>
        <div>
          <div class="text-xs font-semibold uppercase tracking-wide text-[#9499a0]">关于</div>
          <ul class="mt-4 space-y-2 text-sm text-[#61666d]">
            <li><a href="#" @click.prevent="soon()" class="hover:text-[#fb7299]">关于我们</a></li>
            <li><a href="#" @click.prevent="soon()" class="hover:text-[#fb7299]">联系合作</a></li>
            <li><a href="#" @click.prevent="soon()" class="hover:text-[#fb7299]">用户协议</a></li>
          </ul>
        </div>
      </div>
      <div class="border-t border-[#f1f2f3]">
        <div class="mx-auto max-w-6xl px-5 py-5 text-xs text-[#9499a0]">
          © {{ new Date().getFullYear() }} {{ siteName }}. 本站点为演示用途，所有作品与数据均为虚构。
        </div>
      </div>
    </footer>

    <!-- 本地提示 -->
    <Transition
      enter-active-class="transition duration-200"
      enter-from-class="opacity-0 translate-y-2"
      leave-active-class="transition duration-200"
      leave-to-class="opacity-0 translate-y-2"
    >
      <div
        v-if="toastMsg"
        class="fixed bottom-6 left-1/2 z-[60] -translate-x-1/2 rounded-full bg-[#18191c] px-5 py-2.5 text-sm text-white shadow-lg"
      >
        {{ toastMsg }}
      </div>
    </Transition>
  </div>
</template>
