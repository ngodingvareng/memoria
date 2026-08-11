import { ImagePreviewList } from '@/components/image-preview-list';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field';
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
  InputGroupText,
  InputGroupTextarea,
} from '@/components/ui/input-group';
import {
  AddMentionButton,
  MentionAutocompletePopover,
  showMentionShareOffers,
  syncMomentMentionsFromText,
  useInlineMentionAutocomplete,
} from '@/features/mentions';
import { ThreadPickerField } from '@/features/threads';
import { useFormSubmitError } from '@/hooks/use-form-submit-error';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadResponse } from '@/lib/api/generated/models';
import {
  getGetMomentsQueryKey,
  usePostMoments,
  usePostMomentsIdImages,
} from '@/lib/api/generated/moments/moments';
import { queryClient } from '@/lib/query-client';
import { PlusSignIcon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { useForm } from '@tanstack/react-form';
import { useNavigate } from '@tanstack/react-router';
import { useRef, useState } from 'react';
import * as z from 'zod';
import { ColorSwatchPicker } from '../../../components/color-swatch-picker';
import { toDatetimeLocalValue, toRFC3339WithOffset } from '../lib/occurred-at';

const formSchema = z.object({
  threadId: z.string().min(1, 'Please choose a thread.'),
  occurredAt: z.string().min(1, 'Date and time are required.'),
  note: z.string().max(10000, 'Note must be at most 10000 characters.'),
  colorHex: z.string(),
  images: z.array(z.instanceof(File)),
});

export function CreateMomentForm() {
  const navigate = useNavigate();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const noteTextareaRef = useRef<HTMLTextAreaElement>(null);
  const noteChangeRef = useRef<(value: string) => void>(() => {});
  const { error: submitError, guard } = useFormSubmitError();
  const [selectedThread, setSelectedThread] =
    useState<GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadResponse | null>(
      null
    );
  const createMoment = usePostMoments();
  const uploadImage = usePostMomentsIdImages();
  const mentionAutocomplete = useInlineMentionAutocomplete({
    onChange: (value) => noteChangeRef.current(value),
    textareaRef: noteTextareaRef,
  });

  const form = useForm({
    defaultValues: {
      threadId: '',
      occurredAt: toDatetimeLocalValue(new Date()),
      note: '',
      colorHex: '',
      images: [] as File[],
    },
    validators: {
      onSubmit: formSchema,
    },
    onSubmit: ({ value }) =>
      guard(async () => {
        const { occurredAt, offsetMinutes } = toRFC3339WithOffset(
          value.occurredAt
        );
        const moment = await createMoment.mutateAsync({
          data: {
            thread_id: value.threadId,
            occurred_at: occurredAt,
            occurred_utc_offset_minutes: offsetMinutes,
            note: value.note || undefined,
            color_hex: value.colorHex || undefined,
          },
        });

        if (value.images.length > 0 && moment.id) {
          await Promise.all(
            value.images.map((image) =>
              uploadImage.mutateAsync({
                id: moment.id!,
                data: { image },
              })
            )
          );
        }

        if (moment.id && value.note) {
          const { shareOffers } = await syncMomentMentionsFromText(
            moment.id,
            value.note
          );
          if (shareOffers.length > 0) {
            showMentionShareOffers(moment.id, shareOffers);
          }
        }

        await queryClient.invalidateQueries({
          queryKey: getGetMomentsQueryKey(),
        });
        navigate({ to: '/' });
      }),
  });

  return (
    <form
      id="create-moment-form"
      onSubmit={(e) => {
        e.preventDefault();
        form.handleSubmit();
      }}
    >
      <FieldGroup>
        <form.Field name="threadId">
          {(field) => {
            const isInvalid =
              field.state.meta.isTouched && !field.state.meta.isValid;
            return (
              <Field data-invalid={isInvalid}>
                <FieldLabel htmlFor={field.name}>Thread</FieldLabel>
                <ThreadPickerField
                  selectedThread={selectedThread}
                  onSelect={(thread) => {
                    setSelectedThread(thread);
                    field.handleChange(thread.id!);
                  }}
                  onBlur={field.handleBlur}
                  className="w-full justify-between"
                />
                {isInvalid && <FieldError errors={field.state.meta.errors} />}
              </Field>
            );
          }}
        </form.Field>

        <form.Field name="occurredAt">
          {(field) => {
            const isInvalid =
              field.state.meta.isTouched && !field.state.meta.isValid;
            return (
              <Field data-invalid={isInvalid}>
                <FieldLabel htmlFor={field.name}>When</FieldLabel>
                <InputGroupInput
                  id={field.name}
                  name={field.name}
                  type="datetime-local"
                  value={field.state.value}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                  aria-invalid={isInvalid}
                />
                {isInvalid && <FieldError errors={field.state.meta.errors} />}
              </Field>
            );
          }}
        </form.Field>

        <form.Field name="note">
          {(field) => {
            noteChangeRef.current = field.handleChange;
            const isInvalid =
              field.state.meta.isTouched && !field.state.meta.isValid;
            return (
              <Field data-invalid={isInvalid}>
                <FieldLabel htmlFor={field.name} hidden>
                  Note
                </FieldLabel>
                <InputGroup>
                  <InputGroupAddon align="block-start">
                    <InputGroupText>Note</InputGroupText>

                    <InputGroupText className="text-muted-foreground ml-auto text-xs tabular-nums">
                      {field.state.value.length}/10000
                    </InputGroupText>
                  </InputGroupAddon>
                  <InputGroupTextarea
                    ref={noteTextareaRef}
                    id={field.name}
                    name={field.name}
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={mentionAutocomplete.handleChange}
                    onKeyDown={mentionAutocomplete.handleKeyDown}
                    onClick={mentionAutocomplete.handleSelectionChange}
                    onKeyUp={mentionAutocomplete.handleSelectionChange}
                    aria-invalid={isInvalid}
                    placeholder="Watching movies"
                    autoComplete="off"
                    className="min-h-40 text-lg!"
                  />
                </InputGroup>
                <MentionAutocompletePopover
                  anchorRef={noteTextareaRef}
                  open={mentionAutocomplete.isOpen}
                  onOpenChange={(open) =>
                    !open && mentionAutocomplete.closeSuggestions()
                  }
                  users={mentionAutocomplete.suggestions}
                  activeIndex={mentionAutocomplete.activeIndex}
                  isLoading={mentionAutocomplete.isLoading}
                  onSelect={mentionAutocomplete.handleSelect}
                />
                {isInvalid && <FieldError errors={field.state.meta.errors} />}
              </Field>
            );
          }}
        </form.Field>

        <form.Field name="colorHex">
          {(field) => (
            <Field>
              <FieldLabel>Color</FieldLabel>
              <ColorSwatchPicker
                value={field.state.value}
                onChange={field.handleChange}
              />
            </Field>
          )}
        </form.Field>

        <form.Field name="images">
          {(field) => (
            <Field>
              <FieldLabel>Images</FieldLabel>
              <div className="flex flex-col gap-2">
                <ImagePreviewList
                  images={field.state.value}
                  onRemove={(index) =>
                    field.handleChange(
                      field.state.value.filter((_, i) => i !== index)
                    )
                  }
                />
                <div className="flex gap-2">
                  <Button
                    type="button"
                    variant="secondary"
                    size="sm"
                    onClick={() => fileInputRef.current?.click()}
                  >
                    <HugeiconsIcon strokeWidth={2.5} icon={PlusSignIcon} />
                    Add images
                  </Button>
                  <AddMentionButton
                    onInsert={mentionAutocomplete.insertAtCursor}
                  />
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept="image/*"
                    multiple
                    className="hidden"
                    onChange={(e) => {
                      const files = Array.from(e.target.files ?? []);
                      e.target.value = '';
                      if (files.length > 0) {
                        field.handleChange([...field.state.value, ...files]);
                      }
                    }}
                  />
                </div>
              </div>
            </Field>
          )}
        </form.Field>

        {submitError && (
          <Alert variant="destructive">
            <AlertDescription>{submitError}</AlertDescription>
          </Alert>
        )}

        <Field orientation="horizontal" className="justify-end">
          <form.Subscribe
            selector={(state) => [state.isDefaultValue, state.isSubmitting]}
          >
            {([isDefaultValue, isSubmitting]) => (
              <Button
                type="submit"
                form="create-moment-form"
                disabled={isDefaultValue || isSubmitting}
              >
                Create
              </Button>
            )}
          </form.Subscribe>
        </Field>
      </FieldGroup>
    </form>
  );
}
