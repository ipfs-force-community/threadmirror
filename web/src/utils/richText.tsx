import React from 'react';
import type { TweetEntities, NoteTweetRichText, Url, UserMention, Hashtag, Tweet } from '@client/models';

interface Range {
  start: number;
  end: number;
  type: 'bold' | 'italic' | 'link' | 'mention' | 'hashtag';
  data?: any;
}

/**
 * Substring function that works with Unicode code points (characters) instead of UTF-16 code units.
 * This mimics Go's rune-based string slicing to ensure compatibility with backend entity indices.
 */
function unicodeAwareSlice(text: string, start: number, end: number): string {
  if (start < 0 || end < start) {
    return '';
  }
  
  // Convert string to array of characters (code points)
  const chars = Array.from(text);
  
  if (start >= chars.length) {
    return '';
  }
  
  if (end > chars.length) {
    end = chars.length;
  }
  
  return chars.slice(start, end).join('');
}

function buildEntityRanges(entities?: TweetEntities): Range[] {
  const ranges: Range[] = [];
  if (!entities) return ranges;
  // URLs
  entities.urls?.forEach((u: Url) => {
    if (u.indices?.length === 2) {
      ranges.push({ start: u.indices[0], end: u.indices[1], type: 'link', data: u.expandedUrl || u.url });
    }
  });
  // Mentions
  entities.userMentions?.forEach((m: UserMention) => {
    if (m.indices?.length === 2) {
      ranges.push({ start: m.indices[0], end: m.indices[1], type: 'mention', data: m.screenName });
    }
  });
  // Hashtags
  entities.hashtags?.forEach((h: Hashtag) => {
    if (h.indices?.length === 2) {
      ranges.push({ start: h.indices[0], end: h.indices[1], type: 'hashtag', data: h.text });
    }
  });
  return ranges;
}

function buildRichTextRanges(rich?: NoteTweetRichText): Range[] {
  const ranges: Range[] = [];
  if (!rich) return ranges;
  rich.richtextTags?.forEach(tag => {
    tag.richtextTypes.forEach(t => {
      const low = t.toLowerCase();
      if (low === 'bold' || low === 'italic') {
        ranges.push({ start: tag.fromIndex, end: tag.toIndex, type: low as 'bold' | 'italic' });
      }
    });
  });
  return ranges;
}

function sortRanges(ranges: Range[]): Range[] {
  return ranges.sort((a, b) => {
    if (a.start === b.start) {
      return b.end - a.end; // longer first
    }
    return a.start - b.start;
  });
}

function decodeHtml(html: string): string {
  if (!html) return html;
  const txt = document.createElement('textarea');
  txt.innerHTML = html;
  return txt.value;
}

export function renderTweetContent(text: string, entities?: TweetEntities, rich?: NoteTweetRichText): React.ReactNode {
  const entityRanges = buildEntityRanges(entities);
  const richRanges = buildRichTextRanges(rich);
  const allRanges = sortRanges([...entityRanges, ...richRanges]);

  const result: React.ReactNode[] = [];
  let cursor = 0;

  const pushText = (t: string) => {
    if (!t) return;
    result.push(decodeHtml(t));
  };

  allRanges.forEach((range, idx) => {
    if (range.start < cursor) {
      return; // overlapping already handled
    }
    // plain text before range - use Unicode-aware slicing
    pushText(unicodeAwareSlice(text, cursor, range.start));
    const sliceRaw = unicodeAwareSlice(text, range.start, range.end);

    const slice = decodeHtml(sliceRaw);
    let node: React.ReactNode = slice;
    switch (range.type) {
      case 'link':
        node = (
          <a key={`link-${idx}-${range.start}`} href={range.data} target="_blank" rel="noopener noreferrer">
            {slice}
          </a>
        );
        break;
      case 'mention':
        node = (
          <a key={`mention-${idx}-${range.start}`} href={`https://x.com/${range.data}`} target="_blank" rel="noopener noreferrer">
            @{range.data}
          </a>
        );
        break;
      case 'hashtag':
        node = (
          <a key={`hashtag-${idx}-${range.start}`} href={`https://x.com/hashtag/${range.data}`} target="_blank" rel="noopener noreferrer">
            #{range.data}
          </a>
        );
        break;
      case 'bold':
        node = <strong key={`b-${idx}-${range.start}`}>{slice}</strong>;
        break;
      case 'italic':
        node = <em key={`i-${idx}-${range.start}`}>{slice}</em>;
        break;
      default:
        break;
    }
    result.push(node);
    cursor = range.end;
  });

  // remaining text - use Unicode-aware slicing
  pushText(unicodeAwareSlice(text, cursor, Array.from(text).length));

  return result;
}

/**
 * Gets the actual displayable text from a tweet, excluding media URLs and other entities.
 * Uses display_text_range to determine what text should be shown to users.
 */
export function getDisplayableText(tweet: Tweet): string {
  if (!tweet.displayTextRange || tweet.displayTextRange.length !== 2) {
    return tweet.text;
  }
  
  const [start, end] = tweet.displayTextRange;
  return unicodeAwareSlice(tweet.text, start, end);
}
