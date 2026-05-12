export const VERSION = '1.0.0';
export const ID_PREFIX = 'ps-ann-';
export const HOST_ID = 'pageshelf-annotation-host';
export const ANNOTATION_SCHEMA = 'pageshelf.annotation.v1';
export const EXPORT_SCHEMA = 'pageshelf.annotations.v1';

interface StorageKeyOptions {
  origin?: string;
  pathname?: string;
}

export function annotationStorageKey({ origin = location.origin, pathname = location.pathname }: StorageKeyOptions = {}): string {
  return `pageshelf:annotations:v1:${origin}${pathname}`;
}

export function createAnnotationId(): string {
  return `psa_${Date.now().toString(36)}${Math.random().toString(36).slice(2)}`;
}
