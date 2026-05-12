export type AnnotationStatus = 'open' | 'resolved';

export interface Anchor {
  quote: string;
  prefix?: string;
  suffix?: string;
  heading: string;
  path: string;
}

export interface Annotation extends Anchor {
  schema: 'pageshelf.annotation.v1';
  id: string;
  createdAt: string;
  updatedAt: string;
  status: AnnotationStatus;
  comment: string;
  url: string;
}

export interface ExportData {
  schema: 'pageshelf.annotations.v1';
  version: string;
  url: string;
  title: string;
  exportedAt: string;
  annotations: Annotation[];
}

export interface AnnotationActions {
  edit(annotation: Annotation): void;
  toggleResolved(annotation: Annotation): void;
  delete(annotation: Annotation): void;
}

export interface AnnotationElements {
  open: HTMLButtonElement;
  close: HTMLButtonElement;
  panel: HTMLElement;
  comment: HTMLTextAreaElement;
  list: HTMLElement;
  count: HTMLElement;
  toast: HTMLElement;
  add: HTMLButtonElement;
  pick: HTMLButtonElement;
  copy: HTMLButtonElement;
  copyJson: HTMLButtonElement;
  exportJson: HTMLButtonElement;
  clear: HTMLButtonElement;
}
