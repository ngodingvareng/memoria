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
        <div className="grid grid-cols-2 gap-2 max-h-[70vh] overflow-y-auto">
          {images.map((image, index) => (
            <img
              key={image + index}
              src={image}
              alt={`Moment image ${index + 1}`}
              className="w-full aspect-square rounded-xl object-cover"
            />
          ))}
        </div>
      </DialogContent>
    </Dialog>
  );
}
