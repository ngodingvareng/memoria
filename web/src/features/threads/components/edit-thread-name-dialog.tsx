import { Alert, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { useFormSubmitError } from '@/hooks/use-form-submit-error';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadResponse } from '@/lib/api/generated/models';
import {
  getGetThreadsIdQueryKey,
  getGetThreadsQueryKey,
  usePutThreadsId,
} from '@/lib/api/generated/threads/threads';
import { queryClient } from '@/lib/query-client';
import { PencilEdit02Icon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import { useForm } from '@tanstack/react-form';
import { useState } from 'react';
import * as z from 'zod';

interface EditThreadNameDialogProps {
  threadId: string;
  thread: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadResponse;
}

const formSchema = z.object({
  name: z
    .string()
    .min(1, 'Name must be at least 1 character.')
    .max(255, 'Name must be at most 255 characters.'),
});

export function EditThreadNameDialog({
  threadId,
  thread,
}: EditThreadNameDialogProps) {
  const [open, setOpen] = useState(false);
  const {
    error: submitError,
    guard,
    clear: clearSubmitError,
  } = useFormSubmitError();
  const updateThread = usePutThreadsId();

  const form = useForm({
    defaultValues: { name: thread.name ?? '' },
    validators: { onSubmit: formSchema },
    onSubmit: ({ value }) =>
      guard(async () => {
        await updateThread.mutateAsync({
          id: threadId,
          data: {
            name: value.name,
            description: thread.description,
            color_hex: thread.color_hex,
          },
        });
        await Promise.all([
          queryClient.invalidateQueries({
            queryKey: getGetThreadsIdQueryKey(threadId),
          }),
          queryClient.invalidateQueries({
            queryKey: getGetThreadsQueryKey(),
          }),
        ]);
        setOpen(false);
      }),
  });

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        setOpen(next);
        if (next) {
          form.reset();
          clearSubmitError();
        }
      }}
    >
      <DialogTrigger
        render={<Button type="button" variant="ghost" size="icon-sm" />}
      >
        <HugeiconsIcon icon={PencilEdit02Icon} strokeWidth={2} />
        <span className="sr-only">Edit name</span>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit thread name</DialogTitle>
        </DialogHeader>
        <form
          id="edit-thread-name-form"
          onSubmit={(e) => {
            e.preventDefault();
            form.handleSubmit();
          }}
        >
          <FieldGroup>
            <form.Field
              name="name"
              children={(field) => {
                const isInvalid =
                  field.state.meta.isTouched && !field.state.meta.isValid;
                return (
                  <Field data-invalid={isInvalid}>
                    <FieldLabel htmlFor={field.name} hidden>
                      Name
                    </FieldLabel>
                    <Input
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      aria-invalid={isInvalid}
                      autoComplete="off"
                      autoFocus
                    />
                    {isInvalid && (
                      <FieldError errors={field.state.meta.errors} />
                    )}
                  </Field>
                );
              }}
            />

            {submitError && (
              <Alert variant="destructive">
                <AlertDescription>{submitError}</AlertDescription>
              </Alert>
            )}
          </FieldGroup>
        </form>
        <DialogFooter>
          <DialogClose render={<Button type="button" variant="ghost" />}>
            Cancel
          </DialogClose>
          <form.Subscribe
            selector={(state) => [state.isSubmitting, state.isDefaultValue]}
            children={([isSubmitting, isDefaultValue]) => (
              <Button
                type="submit"
                form="edit-thread-name-form"
                disabled={Boolean(isSubmitting) || Boolean(isDefaultValue)}
              >
                Save
              </Button>
            )}
          />
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
