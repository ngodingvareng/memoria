import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';

interface MomentImagesDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  images: string[];
}

export function MomentImagesDialog({
  open,
  onOpenChange,
  images,
}: MomentImagesDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Images</DialogTitle>
        </DialogHeader>
        <div className="grid max-h-[70vh] grid-cols-2 gap-2 overflow-y-auto">
          {images.map((image, index) => (
            <img
              key={image + index}
              src={image}
              alt={`Moment image ${index + 1}`}
              className="aspect-square w-full rounded-xl object-cover"
            />
          ))}
        </div>
      </DialogContent>
    </Dialog>
  );
}
