import { AlbumImageItem } from './album-image-item';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoAlbumImageResponse as AlbumImage } from '@/lib/api/generated/models';

interface AlbumImageListProps {
  images: AlbumImage[];
  onImageClick?: (image: AlbumImage) => void;
}

export function AlbumImageList({ images, onImageClick }: AlbumImageListProps) {
  return (
    <div className="grid grid-cols-3 gap-x-1.5 gap-y-1.5">
      {images.map((image) => (
        <AlbumImageItem
          key={image.id}
          image={image}
          onClick={onImageClick ? () => onImageClick(image) : undefined}
        />
      ))}
    </div>
  );
}
