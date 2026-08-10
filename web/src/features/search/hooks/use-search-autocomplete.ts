import type {
  GithubComNgodingvarengMemoriaInternalDeliveryRestDtoMomentSuggestionResponse,
  GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadSuggestionResponse,
} from '@/lib/api/generated/models';
import { useNavigate } from '@tanstack/react-router';
import * as React from 'react';
import { useSearchSuggestions } from './use-search-suggestions';

type ThreadSuggestion =
  GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadSuggestionResponse;
type MomentSuggestion =
  GithubComNgodingvarengMemoriaInternalDeliveryRestDtoMomentSuggestionResponse;

export type SearchSuggestionItem =
  | { type: 'thread'; data: ThreadSuggestion }
  | { type: 'moment'; data: MomentSuggestion };

// Drives the header search input's live-typing popover: debounced
// suggestions, ArrowUp/Down across the combined Threads+Moments list,
// Enter on a highlighted suggestion navigates to its detail page, Enter
// with nothing highlighted navigates to the full /search results page.
export function useSearchAutocomplete() {
  const [value, setValue] = React.useState('');
  const [isOpen, setIsOpen] = React.useState(false);
  const [activeIndex, setActiveIndex] = React.useState(0);
  const navigate = useNavigate();

  const { threads, moments, isLoading } = useSearchSuggestions(value);

  const items: SearchSuggestionItem[] = React.useMemo(
    () => [
      ...threads.map((data) => ({ type: 'thread' as const, data })),
      ...moments.map((data) => ({ type: 'moment' as const, data })),
    ],
    [threads, moments]
  );

  const goToSearchPage = React.useCallback(
    (query: string) => {
      setIsOpen(false);
      navigate({ to: '/search', search: { q: query } });
    },
    [navigate]
  );

  const handleSelect = React.useCallback(
    (item: SearchSuggestionItem) => {
      setIsOpen(false);
      if (item.type === 'thread') {
        navigate({ to: '/thread/$id', params: { id: item.data.id! } });
      } else {
        navigate({ to: '/moment/$id', params: { id: item.data.id! } });
      }
    },
    [navigate]
  );

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const nextValue = event.target.value;
    setValue(nextValue);
    setIsOpen(nextValue.trim().length > 0);
    setActiveIndex(0);
  };

  const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    const trimmed = value.trim();

    if (!isOpen) {
      if (event.key === 'Enter' && trimmed.length > 0) {
        event.preventDefault();
        goToSearchPage(trimmed);
      }
      return;
    }

    switch (event.key) {
      case 'ArrowDown':
        event.preventDefault();
        if (items.length > 0) setActiveIndex((i) => (i + 1) % items.length);
        break;
      case 'ArrowUp':
        event.preventDefault();
        if (items.length > 0) {
          setActiveIndex((i) => (i - 1 + items.length) % items.length);
        }
        break;
      case 'Enter':
        event.preventDefault();
        if (items.length > 0 && items[activeIndex]) {
          handleSelect(items[activeIndex]);
        } else if (trimmed.length > 0) {
          goToSearchPage(trimmed);
        }
        break;
      case 'Escape':
        setIsOpen(false);
        break;
    }
  };

  return {
    value,
    isOpen,
    setIsOpen,
    items,
    isLoading,
    activeIndex,
    handleChange,
    handleKeyDown,
    handleSelect,
  };
}
