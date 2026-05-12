import { ANNOTATION_SCHEMA, createAnnotationId } from './constants';
import type { Annotation, AnnotationStatus } from './types';

type RawAnnotation = Partial<Record<keyof Annotation, unknown>>;

function textValue(value: unknown, fallback = ''): string {
  return typeof value === 'string' ? value : fallback;
}

function statusValue(value: unknown): AnnotationStatus {
  return value === 'resolved' ? 'resolved' : 'open';
}

function normalizeAnnotation(annotation: RawAnnotation, defaultUrl = location.href): Annotation {
  const createdAt = textValue(annotation.createdAt, new Date().toISOString());

  return {
    schema: ANNOTATION_SCHEMA,
    id: textValue(annotation.id, createAnnotationId()),
    createdAt,
    updatedAt: textValue(annotation.updatedAt, createdAt),
    status: statusValue(annotation.status),
    comment: textValue(annotation.comment),
    quote: textValue(annotation.quote),
    prefix: textValue(annotation.prefix),
    suffix: textValue(annotation.suffix),
    heading: textValue(annotation.heading),
    path: textValue(annotation.path),
    url: textValue(annotation.url, defaultUrl),
  };
}

export function loadAnnotations(storageKey: string, storage: Storage = localStorage): Annotation[] {
  let parsed: unknown;

  try {
    parsed = JSON.parse(storage.getItem(storageKey) || '[]');
  } catch {
    parsed = [];
  }

  if (!Array.isArray(parsed)) {
    return [];
  }

  return parsed.map((annotation: unknown) => normalizeAnnotation(annotation as RawAnnotation));
}

export function saveAnnotations(storageKey: string, annotations: Annotation[], storage: Storage = localStorage): void {
  storage.setItem(storageKey, JSON.stringify(annotations));
}
