import type { ReactNode } from 'react';

export interface Story {
  user: {
    id: number;
    name: string;
    username: string;
    imageSrc: string;
    imageAlt: string;
  };
  id: number;
  date: string;
  time: string;
  tags: string[];
  content: ReactNode;
  color?: string;
}
