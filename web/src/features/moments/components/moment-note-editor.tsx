import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import {
  AddMentionButton,
  MentionAutocompletePopover,
  useInlineMentionAutocomplete,
} from '@/features/mentions';
import React from 'react';

interface MomentNoteEditorProps {
  content: string;
  onCancel: () => void;
  onSave: (note: string) => void | Promise<void>;
}

export function MomentNoteEditor({
  content,
  onCancel,
  onSave,
}: MomentNoteEditorProps) {
  const [draft, setDraft] = React.useState(content);
  const [isSaving, setIsSaving] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const textareaRef = React.useRef<HTMLTextAreaElement>(null);
  const mentionAutocomplete = useInlineMentionAutocomplete({
    onChange: setDraft,
    textareaRef,
  });

  const handleSave = async () => {
    setError(null);
    setIsSaving(true);
    try {
      await onSave(draft);
    } catch {
      setError('Failed to save. Please try again.');
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="flex flex-col gap-2">
      <Textarea
        ref={textareaRef}
        value={draft}
        onChange={mentionAutocomplete.handleChange}
        onKeyDown={mentionAutocomplete.handleKeyDown}
        onClick={mentionAutocomplete.handleSelectionChange}
        onKeyUp={mentionAutocomplete.handleSelectionChange}
        maxLength={10000}
        autoFocus
        className="text-base min-h-24"
        disabled={isSaving}
      />
      <MentionAutocompletePopover
        anchorRef={textareaRef}
        open={mentionAutocomplete.isOpen}
        onOpenChange={(open) => !open && mentionAutocomplete.closeSuggestions()}
        users={mentionAutocomplete.suggestions}
        activeIndex={mentionAutocomplete.activeIndex}
        isLoading={mentionAutocomplete.isLoading}
        onSelect={mentionAutocomplete.handleSelect}
      />
      <div className="flex gap-2 justify-between">
        <AddMentionButton onInsert={mentionAutocomplete.insertAtCursor} />
        <div className="flex gap-2">
          <Button
            size="sm"
            variant="ghost"
            disabled={isSaving}
            onClick={onCancel}
          >
            Cancel
          </Button>
          <Button size="sm" disabled={isSaving} onClick={handleSave}>
            Save
          </Button>
        </div>
      </div>
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
    </div>
  );
}
