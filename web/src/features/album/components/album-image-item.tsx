import { CambioImage } from '@/components/ui/cambio-image';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoAlbumImageResponse as AlbumImage } from '@/lib/api/generated/models';

interface AlbumImageItemProps {
  image: AlbumImage;
}

export function AlbumImageItem({ image }: AlbumImageItemProps) {
  return (
    <div className="relative">
      <CambioImage
        src={image.url ?? ''}
        alt={image.image_alt ?? ''}
        width={image.width ?? 800}
        height={image.height ?? 600}
        motion="smooth"
        className="rounded-md"
      />
      {image.is_shared && image.shared_by && (
        <span className="absolute bottom-2 left-2 rounded-full bg-black/60 px-2 py-0.5 text-xs text-white">
          from{' '}
          {image.shared_by.username
            ? `@${image.shared_by.username}`
            : 'a circle member'}
        </span>
      )}
    </div>
  );
}
