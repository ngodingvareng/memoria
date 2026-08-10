import { AlbumImageItem } from './album-image-item';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoAlbumImageResponse as AlbumImage } from '@/lib/api/generated/models';

interface AlbumImageListProps {
  images: AlbumImage[];
}

export function AlbumImageList({ images }: AlbumImageListProps) {
  return (
    <div className="grid grid-cols-3 gap-x-1.5 gap-y-1.5">
      {images.map((image) => (
        <AlbumImageItem key={image.id} image={image} />
      ))}
    </div>
  );
}
