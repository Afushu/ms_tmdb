export function looksLikeJsonText(value: string | undefined) {
  const text = (value ?? "").trimStart();
  return text.startsWith("{") || text.startsWith("[");
}

export function formatJsonText(value: string | undefined) {
  const raw = value ?? "";
  if (!raw.trim()) {
    return "(空)";
  }
  if (!looksLikeJsonText(raw)) {
    return raw;
  }

  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return formatPartialJsonText(raw);
  }
}

// 日志正文可能已被后端截断，严格 JSON.parse 失败时仍尽量按 JSON 符号排版。
function formatPartialJsonText(value: string) {
  const source = value.trim();
  let result = "";
  let indent = 0;
  let inString = false;
  let escaped = false;

  const appendNewLine = () => {
    result = result.trimEnd();
    result += `\n${"  ".repeat(Math.max(0, indent))}`;
  };

  for (const char of source) {
    if (inString) {
      result += char;
      if (escaped) {
        escaped = false;
      } else if (char === "\\") {
        escaped = true;
      } else if (char === '"') {
        inString = false;
      }
      continue;
    }

    if (char === '"') {
      inString = true;
      result += char;
      continue;
    }

    if (char === "{" || char === "[") {
      result += char;
      indent += 1;
      appendNewLine();
      continue;
    }

    if (char === "}" || char === "]") {
      indent = Math.max(0, indent - 1);
      result = result.trimEnd();
      result += `\n${"  ".repeat(indent)}${char}`;
      continue;
    }

    if (char === ",") {
      result += char;
      appendNewLine();
      continue;
    }

    if (char === ":") {
      result += ": ";
      continue;
    }

    if (/\s/.test(char)) {
      continue;
    }

    result += char;
  }

  return result.trimEnd();
}
