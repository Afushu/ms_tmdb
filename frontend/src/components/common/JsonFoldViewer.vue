<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { formatJsonText, looksLikeJsonText } from "@/utils/jsonText";

type FoldRange = {
  start: number;
  end: number;
  childLines: number;
  closeText: string;
  depth: number;
};

const props = defineProps<{
  value?: string;
}>();

const collapsedLines = ref<Set<number>>(new Set());

const parseResult = computed(() => {
  const raw = props.value ?? "";
  if (!raw.trim()) {
    return { json: false, text: "(空)", lines: ["(空)"] };
  }

  try {
    const parsed = JSON.parse(raw) as unknown;
    return {
      json: true,
      text: "",
      lines: JSON.stringify(parsed, null, 2).split("\n"),
    };
  } catch {
    const text = formatJsonText(raw);
    return { json: looksLikeJsonText(raw), text, lines: text.split("\n") };
  }
});

const foldRanges = computed(() => buildFoldRanges(parseResult.value.lines));
const foldRangeMap = computed(() => new Map(foldRanges.value.map((range) => [range.start, range])));
const hiddenLines = computed(() => {
  const hidden = new Set<number>();
  for (const start of collapsedLines.value) {
    const range = foldRangeMap.value.get(start);
    if (!range) {
      continue;
    }
    for (let index = range.start + 1; index <= range.end; index += 1) {
      hidden.add(index);
    }
  }
  return hidden;
});

const visibleRows = computed(() =>
  parseResult.value.lines
    .map((line, index) => {
      const range = foldRangeMap.value.get(index);
      const collapsed = collapsedLines.value.has(index);
      return {
        index,
        line: collapsed && range ? collapsedLineText(line, range) : line,
        foldable: Boolean(range),
        collapsed,
      };
    })
    .filter((row) => !hiddenLines.value.has(row.index)),
);

watch(
  () => props.value,
  () => {
    collapsedLines.value = defaultCollapsedLines(foldRanges.value);
  },
  { immediate: true },
);

function lineIndent(line: string) {
  return line.match(/^\s*/)?.[0].length ?? 0;
}

function foldCloseChar(line: string) {
  const trimmed = line.trimEnd();
  if (trimmed.endsWith("{")) return "}";
  if (trimmed.endsWith("[")) return "]";
  return "";
}

function buildFoldRanges(lines: string[]) {
  const stack: Array<{ index: number; indent: number; closeChar: string }> = [];
  const ranges: FoldRange[] = [];

  lines.forEach((line, index) => {
    const indent = lineIndent(line);
    const trimmed = line.trimStart();
    const firstChar = trimmed[0];

    if ((firstChar === "}" || firstChar === "]") && stack.length > 0) {
      const top = stack[stack.length - 1];
      if (top.closeChar === firstChar && top.indent === indent) {
        stack.pop();
        if (index > top.index + 1) {
          ranges.push({
            start: top.index,
            end: index,
            childLines: index - top.index - 1,
            closeText: trimmed,
            depth: Math.floor(top.indent / 2),
          });
        }
      }
    }

    const closeChar = foldCloseChar(line);
    if (closeChar) {
      stack.push({ index, indent, closeChar });
    }
  });

  return ranges;
}

function defaultCollapsedLines(ranges: FoldRange[]) {
  return new Set(ranges.filter((range) => range.start > 0 && range.childLines > 8).map((range) => range.start));
}

function collapsedLineText(line: string, range: FoldRange) {
  return `${line.trimEnd()} ... ${range.childLines} 行 ${range.closeText}`;
}

function toggleLine(index: number) {
  const next = new Set(collapsedLines.value);
  if (next.has(index)) {
    next.delete(index);
  } else {
    next.add(index);
  }
  collapsedLines.value = next;
}

function collapseAll() {
  collapsedLines.value = new Set(foldRanges.value.map((range) => range.start));
}

function expandAll() {
  collapsedLines.value = new Set();
}
</script>

<template>
  <div v-if="parseResult.json" class="json-fold-viewer">
    <div v-if="foldRanges.length > 0" class="json-fold-toolbar">
      <button type="button" @click="collapseAll">全部折叠</button>
      <button type="button" @click="expandAll">全部展开</button>
    </div>

    <div class="json-fold-code">
      <div
        v-for="row in visibleRows"
        :key="row.index"
        class="json-fold-row"
        :class="row.collapsed ? 'json-fold-row-collapsed' : ''"
      >
        <span class="json-fold-line-number">{{ row.index + 1 }}</span>
        <button
          v-if="row.foldable"
          type="button"
          class="json-fold-toggle"
          :aria-label="row.collapsed ? '展开代码块' : '折叠代码块'"
          @click="toggleLine(row.index)"
        >
          {{ row.collapsed ? "+" : "-" }}
        </button>
        <span v-else class="json-fold-toggle-spacer"></span>
        <code class="json-fold-line">{{ row.line }}</code>
      </div>
    </div>
  </div>

  <pre v-else class="settings-diff-pre logs-detail-response-body">{{ parseResult.text }}</pre>
</template>
