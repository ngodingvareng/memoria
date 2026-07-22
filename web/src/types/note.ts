import type { ReactNode } from "react";

export interface Note {
  id: number;
  date: string;
  time: string;
  published: boolean;
  tags: string[];
  content: ReactNode;
  color?: string;
}
