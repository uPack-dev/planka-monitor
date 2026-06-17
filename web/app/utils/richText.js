export function formatRichText(value, options = {}) {
  const text = String(value || '')
    .replace(/\r\n?/g, '\n')
    .trim();
  if (!text) return '';

  return text
    .split(/\n{2,}/)
    .map((block) => formatRichTextBlock(block, options))
    .join('');
}

function formatRichTextBlock(block, options) {
  const lines = String(block || '')
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean);

  if (!lines.length) return '';

  const headingMatch = lines[0].match(/^(#{1,6})\s+(.+)$/);

  if (headingMatch) {
    const rest = lines.slice(1).join('\n').trim();
    return [
      `<h4>${formatInlineText(headingMatch[2], options)}</h4>`,
      rest ? formatRichTextBlock(rest, options) : '',
    ].join('');
  }

  const parts = [];
  let paragraph = [];
  let listType = '';
  let listItems = [];

  const flushParagraph = () => {
    if (!paragraph.length) return;
    parts.push(
      `<p>${paragraph
        .map((line) => formatInlineText(line, options))
        .join('<br>')}</p>`,
    );
    paragraph = [];
  };

  const flushList = () => {
    if (!listItems.length) return;
    const tag = listType === 'ol' ? 'ol' : 'ul';
    const itemsHtml = listItems
      .map(
        (item) =>
          `<li>${formatInlineText(item, options).replace(/\n/g, '<br>')}</li>`,
      )
      .join('');
    parts.push(`<${tag}>${itemsHtml}</${tag}>`);
    listType = '';
    listItems = [];
  };

  for (const line of lines) {
    const listItem = parseListLine(line);
    if (listItem) {
      flushParagraph();
      if (listType && listType !== listItem.type) {
        flushList();
      }
      listType = listItem.type;
      listItems.push(listItem.text);
      continue;
    }

    flushList();
    paragraph.push(line);
  }

  flushParagraph();
  flushList();

  return parts.join('');
}

function parseListLine(line) {
  const ordered = line.match(/^\d+[).]\s+(.+)$/);
  if (ordered) {
    return { type: 'ol', text: ordered[1] };
  }

  const unordered = line.match(/^[-*•]\s+(.+)$/);
  if (unordered) {
    return { type: 'ul', text: unordered[1] };
  }

  return null;
}

function formatInlineText(value, options) {
  let html = escapeHtml(value);
  const mentionClass = options.mentionClass || '';

  html = html.replace(/@\[([^\]]+)]\(([^)]+)\)/g, (_match, label) => {
    const classAttr = mentionClass
      ? ` class="${escapeHtml(mentionClass)}"`
      : '';
    return `<span${classAttr}>@${label}</span>`;
  });
  html = html.replace(
    /\[([^\]]+)]\((https?:\/\/[^)\s]+)\)/g,
    (_match, label, url) => linkHtml(url, label),
  );
  html = html.replace(/&lt;(https?:\/\/[^<\s]+)&gt;/g, (_match, url) =>
    linkHtml(url, url),
  );
  html = html.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
  html = html.replace(/`([^`]+)`/g, '<code>$1</code>');

  return html;
}

function linkHtml(url, label) {
  return `<a href="${url}" target="_blank" rel="noreferrer">${label}</a>`;
}

export function escapeHtml(value) {
  return String(value)
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#039;');
}
