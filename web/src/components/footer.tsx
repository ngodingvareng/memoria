import * as React from 'react';
import { Button } from './ui/button';
import Wrapper from './wrapper';

type Theme = 'light' | 'dark' | 'system';

function getSystemTheme(): 'light' | 'dark' {
  return window.matchMedia('(prefers-color-scheme: dark)').matches
    ? 'dark'
    : 'light';
}

function applyTheme(theme: Theme) {
  const root = document.documentElement;
  if (theme === 'system') {
    const systemTheme = getSystemTheme();
    root.classList.toggle('dark', systemTheme === 'dark');
  } else {
    root.classList.toggle('dark', theme === 'dark');
  }
}

function getStoredTheme(): Theme {
  const stored = localStorage.getItem('theme');
  if (stored === 'light' || stored === 'dark' || stored === 'system') {
    return stored;
  }
  return 'system';
}

export default function Footer() {
  const [theme, setTheme] = React.useState<Theme>(() => {
    if (typeof window !== 'undefined') {
      return getStoredTheme();
    }
    return 'system';
  });

  React.useEffect(() => {
    applyTheme(theme);
    localStorage.setItem('theme', theme);
  }, [theme]);

  React.useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handleChange = () => {
      if (theme === 'system') {
        applyTheme('system');
      }
    };
    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, [theme]);

  const cycleTheme = () => {
    setTheme((current) => {
      if (current === 'light') return 'dark';
      if (current === 'dark') return 'system';
      return 'light';
    });
  };

  const themeLabel = theme.charAt(0).toUpperCase() + theme.slice(1);

  return (
    <footer className="flex bg-background">
      <Wrapper>
        <div className="flex items-center gap-4 w-full justify-between">
          <p className="text-muted-foreground text-sm">
            &copy;2025 Micin. All rights reserved.
          </p>
          <Button variant="ghost" size="sm" onClick={cycleTheme}>
            {themeLabel}
          </Button>
        </div>
      </Wrapper>
    </footer>
  );
}
