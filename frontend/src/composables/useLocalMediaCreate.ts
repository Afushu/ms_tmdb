import { computed, ref, type Ref } from "vue";
import { useRouter } from "vue-router";
import {
  createMovie,
  createTV,
  uploadAdminImage,
  type AdminCreateMoviePayload,
  type AdminCreateTVPayload,
} from "@/api/admin";
import { clearMovieCache, getMovieGenreList } from "@/api/movie";
import { clearTVCache, getTVGenreList } from "@/api/tv";
import type {
  LocalMovieCreateForm,
  LocalTVCreateForm,
  MediaTab,
  UploadingKey,
} from "@/components/library/types";
import { movieStatusOptions, tvStatusOptions, tvTypeOptions } from "@/constants/mediaStatus";
import { normalizeGenreOptions, type GenreOption } from "@/utils/mediaNormalizers";

type UseLocalMediaCreateOptions = {
  activeTab: Ref<MediaTab>;
  loadData: () => Promise<void>;
};

export function useLocalMediaCreate(options: UseLocalMediaCreateOptions) {
  const router = useRouter();
  const { activeTab, loadData } = options;

  const createPanelVisible = ref(false);
  const creating = ref(false);
  const createError = ref("");
  const uploadingKey = ref<UploadingKey>("");
  const movieCreateForm = ref<LocalMovieCreateForm>(emptyMovieForm());
  const tvCreateForm = ref<LocalTVCreateForm>(emptyTVForm());
  const movieGenreOptions = ref<GenreOption[]>([]);
  const tvGenreOptions = ref<GenreOption[]>([]);

  const languageOptions = [
    { label: "中文 (zh-CN)", value: "zh-CN" },
    { label: "英语 (en-US)", value: "en-US" },
    { label: "日语 (ja-JP)", value: "ja-JP" },
    { label: "韩语 (ko-KR)", value: "ko-KR" },
  ] as const;

  const createTitle = computed(() => (activeTab.value === "movie" ? "新建本地电影" : "新建本地剧集"));

  function emptyMovieForm(): LocalMovieCreateForm {
    return {
      title: "",
      original_title: "",
      genre_names: [],
      release_date: "",
      status: "Released",
      runtime: "",
      original_language: "zh-CN",
      poster_path: "",
      backdrop_path: "",
      vote_average: "",
      popularity: "",
      overview: "",
    };
  }

  function emptyTVForm(): LocalTVCreateForm {
    return {
      name: "",
      original_name: "",
      genre_names: [],
      first_air_date: "",
      status: "Returning Series",
      type: "Scripted",
      number_of_seasons: "",
      number_of_episodes: "",
      original_language: "zh-CN",
      poster_path: "",
      backdrop_path: "",
      vote_average: "",
      popularity: "",
      overview: "",
    };
  }

  function resetCreateForm() {
    movieCreateForm.value = emptyMovieForm();
    tvCreateForm.value = emptyTVForm();
    createError.value = "";
    uploadingKey.value = "";
  }

  async function loadMovieGenreOptions() {
    try {
      // 创建表单辅助请求：失败静默降级为空选项
      const resp = await getMovieGenreList("zh-CN", { showErrorToast: false });
      movieGenreOptions.value = normalizeGenreOptions(resp.data?.genres);
    } catch {
      movieGenreOptions.value = [];
    }
  }

  async function loadTVGenreOptions() {
    try {
      // 创建表单辅助请求：失败静默降级为空选项
      const resp = await getTVGenreList("zh-CN", { showErrorToast: false });
      tvGenreOptions.value = normalizeGenreOptions(resp.data?.genres);
    } catch {
      tvGenreOptions.value = [];
    }
  }

  function openCreatePanel() {
    createPanelVisible.value = true;
    resetCreateForm();
    if (activeTab.value === "movie") {
      void loadMovieGenreOptions();
    } else {
      void loadTVGenreOptions();
    }
  }

  function closeCreatePanel() {
    createPanelVisible.value = false;
    resetCreateForm();
  }

  /** 外部路由 tab 变更时的副作用：关闭面板但不完整重置表单字段，与原页面行为一致 */
  function onExternalTabChange(tab: MediaTab) {
    createPanelVisible.value = false;
    createError.value = "";
    if (tab === "movie") {
      void loadMovieGenreOptions();
    } else {
      void loadTVGenreOptions();
    }
  }

  function parseOptionalInt(raw: string): number | undefined {
    const text = raw.trim();
    if (!text) return undefined;
    const value = Number(text);
    if (!Number.isFinite(value)) return undefined;
    return Math.trunc(value);
  }

  function parseOptionalFloat(raw: string): number | undefined {
    const text = raw.trim();
    if (!text) return undefined;
    const value = Number(text);
    if (!Number.isFinite(value)) return undefined;
    return value;
  }

  async function uploadCreateImage(mediaType: MediaTab, field: "poster_path" | "backdrop_path", event: Event) {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;

    const key = `${mediaType}_${field}` as UploadingKey;
    uploadingKey.value = key;
    createError.value = "";
    try {
      const resp = await uploadAdminImage(file);
      const path = String(resp.data?.path ?? "").trim();
      if (!path) {
        throw new Error("上传成功但未返回图片路径");
      }
      if (mediaType === "movie") {
        movieCreateForm.value[field] = path;
      } else {
        tvCreateForm.value[field] = path;
      }
    } catch {
      /* handled by global toast */
    } finally {
      uploadingKey.value = "";
      input.value = "";
    }
  }

  async function submitCreate() {
    createError.value = "";

    if (activeTab.value === "movie") {
      const title = movieCreateForm.value.title.trim();
      if (!title) {
        createError.value = "电影标题不能为空";
        return;
      }

      const runtime = parseOptionalInt(movieCreateForm.value.runtime);
      if (movieCreateForm.value.runtime.trim() && runtime === undefined) {
        createError.value = "时长必须是数字";
        return;
      }
      const voteAverage = parseOptionalFloat(movieCreateForm.value.vote_average);
      if (movieCreateForm.value.vote_average.trim() && voteAverage === undefined) {
        createError.value = "评分必须是数字";
        return;
      }
      const popularity = parseOptionalFloat(movieCreateForm.value.popularity);
      if (movieCreateForm.value.popularity.trim() && popularity === undefined) {
        createError.value = "热度必须是数字";
        return;
      }

      const payload: AdminCreateMoviePayload = {
        title,
        original_title: movieCreateForm.value.original_title.trim(),
        release_date: movieCreateForm.value.release_date.trim(),
        status: movieCreateForm.value.status.trim(),
        original_language: movieCreateForm.value.original_language.trim(),
        poster_path: movieCreateForm.value.poster_path.trim(),
        backdrop_path: movieCreateForm.value.backdrop_path.trim(),
        overview: movieCreateForm.value.overview.trim(),
        genre_names: movieCreateForm.value.genre_names,
      };
      if (runtime !== undefined) payload.runtime = runtime;
      if (voteAverage !== undefined) payload.vote_average = voteAverage;
      if (popularity !== undefined) payload.popularity = popularity;

      creating.value = true;
      try {
        const resp = await createMovie(payload);
        const createdID = Number(resp.data?.tmdb_id);
        if (!Number.isInteger(createdID)) {
          throw new Error("创建成功但未返回有效 ID");
        }
        clearMovieCache(createdID);
        closeCreatePanel();
        await loadData();
        await router.push(`/movie/${createdID}`);
      } catch {
        /* handled by global toast */
      } finally {
        creating.value = false;
      }
      return;
    }

    const name = tvCreateForm.value.name.trim();
    if (!name) {
      createError.value = "剧集名称不能为空";
      return;
    }

    const seasons = parseOptionalInt(tvCreateForm.value.number_of_seasons);
    if (tvCreateForm.value.number_of_seasons.trim() && seasons === undefined) {
      createError.value = "季数必须是数字";
      return;
    }
    const episodes = parseOptionalInt(tvCreateForm.value.number_of_episodes);
    if (tvCreateForm.value.number_of_episodes.trim() && episodes === undefined) {
      createError.value = "集数必须是数字";
      return;
    }
    const voteAverage = parseOptionalFloat(tvCreateForm.value.vote_average);
    if (tvCreateForm.value.vote_average.trim() && voteAverage === undefined) {
      createError.value = "评分必须是数字";
      return;
    }
    const popularity = parseOptionalFloat(tvCreateForm.value.popularity);
    if (tvCreateForm.value.popularity.trim() && popularity === undefined) {
      createError.value = "热度必须是数字";
      return;
    }

    const payload: AdminCreateTVPayload = {
      name,
      original_name: tvCreateForm.value.original_name.trim(),
      first_air_date: tvCreateForm.value.first_air_date.trim(),
      status: tvCreateForm.value.status.trim(),
      type: tvCreateForm.value.type.trim(),
      original_language: tvCreateForm.value.original_language.trim(),
      poster_path: tvCreateForm.value.poster_path.trim(),
      backdrop_path: tvCreateForm.value.backdrop_path.trim(),
      overview: tvCreateForm.value.overview.trim(),
      genre_names: tvCreateForm.value.genre_names,
    };
    if (seasons !== undefined) payload.number_of_seasons = seasons;
    if (episodes !== undefined) payload.number_of_episodes = episodes;
    if (voteAverage !== undefined) payload.vote_average = voteAverage;
    if (popularity !== undefined) payload.popularity = popularity;

    creating.value = true;
    try {
      const resp = await createTV(payload);
      const createdID = Number(resp.data?.tmdb_id);
      if (!Number.isInteger(createdID)) {
        throw new Error("创建成功但未返回有效 ID");
      }
      clearTVCache(createdID);
      closeCreatePanel();
      await loadData();
      await router.push(`/tv/${createdID}`);
    } catch {
      /* handled by global toast */
    } finally {
      creating.value = false;
    }
  }

  return {
    createPanelVisible,
    creating,
    createError,
    uploadingKey,
    movieCreateForm,
    tvCreateForm,
    movieGenreOptions,
    tvGenreOptions,
    languageOptions,
    movieStatusOptions,
    tvStatusOptions,
    tvTypeOptions,
    createTitle,
    openCreatePanel,
    closeCreatePanel,
    onExternalTabChange,
    uploadCreateImage,
    submitCreate,
  };
}
