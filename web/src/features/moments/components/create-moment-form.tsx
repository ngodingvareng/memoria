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
import { ApiError } from '@/lib/api-client';
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
import { ColorSwatchPicker } from './color-swatch-picker';
import { ImagePreviewList } from './image-preview-list';
import { toDatetimeLocalValue, toRFC3339WithOffset } from '../lib/occurred-at';

const formSchema = z.object({
  occurredAt: z.string().min(1, 'Date and time are required.'),
  note: z.string().max(10000, 'Note must be at most 10000 characters.'),
  colorHex: z.string(),
  images: z.array(z.instanceof(File)),
});

export function CreateMomentForm() {
  const navigate = useNavigate();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const createMoment = usePostMoments();
  const uploadImage = usePostMomentsIdImages();

  const form = useForm({
    defaultValues: {
      occurredAt: toDatetimeLocalValue(new Date()),
      note: '',
      colorHex: '',
      images: [] as File[],
    },
    validators: {
      onSubmit: formSchema,
    },
    onSubmit: async ({ value }) => {
      setSubmitError(null);
      try {
        const { occurredAt, offsetMinutes } = toRFC3339WithOffset(
          value.occurredAt
        );
        const moment = await createMoment.mutateAsync({
          data: {
            occurred_at: occurredAt,
            occurred_utc_offset_minutes: offsetMinutes,
            note: value.note || undefined,
            color_hex: value.colorHex || undefined,
          },
        });

        if (value.images.length > 0 && moment.id) {
          await Promise.all(
            value.images.map((image) =>
              uploadImage.mutateAsync({ id: moment.id!, data: { image } })
            )
          );
        }

        await queryClient.invalidateQueries({
          queryKey: getGetMomentsQueryKey(),
        });
        navigate({ to: '/' });
      } catch (err) {
        if (err instanceof ApiError) {
          setSubmitError(
            err.fieldErrors?.map((e) => e.message).join(' ') ?? err.message
          );
        } else {
          setSubmitError('Something went wrong. Please try again.');
        }
      }
    },
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
        <form.Field
          name="occurredAt"
          children={(field) => {
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
        />

        <form.Field
          name="note"
          children={(field) => {
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

                    <InputGroupText className="ml-auto tabular-nums text-xs text-muted-foreground">
                      {field.state.value.length}/10000
                    </InputGroupText>
                  </InputGroupAddon>
                  <InputGroupTextarea
                    id={field.name}
                    name={field.name}
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    aria-invalid={isInvalid}
                    placeholder="Watching movies"
                    autoComplete="off"
                    className="text-lg! min-h-40"
                  />
                </InputGroup>
                {isInvalid && <FieldError errors={field.state.meta.errors} />}
              </Field>
            );
          }}
        />

        <form.Field
          name="colorHex"
          children={(field) => (
            <Field>
              <FieldLabel>Color</FieldLabel>
              <ColorSwatchPicker
                value={field.state.value}
                onChange={field.handleChange}
              />
            </Field>
          )}
        />

        <form.Field
          name="images"
          children={(field) => (
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
                <div>
                  <Button
                    type="button"
                    variant="secondary"
                    size="sm"
                    onClick={() => fileInputRef.current?.click()}
                  >
                    <HugeiconsIcon strokeWidth={2.5} icon={PlusSignIcon} />
                    Add images
                  </Button>
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
        />

        {submitError && (
          <Alert variant="destructive">
            <AlertDescription>{submitError}</AlertDescription>
          </Alert>
        )}

        <Field orientation="horizontal" className="justify-end">
          <form.Subscribe
            selector={(state) => [state.isDefaultValue, state.isSubmitting]}
            children={([isDefaultValue, isSubmitting]) => (
              <Button
                type="submit"
                form="create-moment-form"
                disabled={isDefaultValue || isSubmitting}
              >
                Create
              </Button>
            )}
          />
        </Field>
      </FieldGroup>
    </form>
  );
}
