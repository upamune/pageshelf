import { EXPORT_SCHEMA, VERSION } from './constants';
import type { Annotation, ExportData } from './types';

export function annotationMarkdown(annotation: Annotation): string {
  return `- [${annotation.status || 'open'}] ${annotation.heading || annotation.path || 'Document'}\n  Quote: "${annotation.quote || ''}"\n  Comment: ${annotation.comment}`;
}

export async function copyText(text: string, onCopied?: () => void): Promise<void> {
  try {
    await navigator.clipboard.writeText(text);
  } catch {
    const textarea = document.createElement('textarea');
    textarea.value = text;
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand('copy');
    textarea.remove();
  }

  onCopied?.();
}

export function exportData(annotations: Annotation[]): ExportData {
  return {
    schema: EXPORT_SCHEMA,
    version: VERSION,
    url: location.href,
    title: document.title,
    exportedAt: new Date().toISOString(),
    annotations,
  };
}

export function downloadJson(filename: string, data: ExportData): void {
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');

  link.href = url;
  link.download = filename;
  link.click();

  setTimeout(() => URL.revokeObjectURL(url), 1000);
}
