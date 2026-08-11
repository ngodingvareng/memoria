import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoAlbumImageResponse as AlbumImage } from '@/lib/api/generated/models';

interface AlbumImageItemProps {
  image: AlbumImage;
  onClick?: () => void;
}

export function AlbumImageItem({ image, onClick }: AlbumImageItemProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="focus-visible:ring-ring/50 relative cursor-pointer rounded-md text-left outline-none focus-visible:ring-[3px]"
    >
      <img
        src={image.url ?? ''}
        alt={image.image_alt ?? ''}
        width={image.width ?? 800}
        height={image.height ?? 600}
        loading="lazy"
        className="block h-auto w-full rounded-md"
      />
      {image.is_shared && image.shared_by && (
        <span className="absolute bottom-2 left-2 rounded-full bg-black/60 px-2 py-0.5 text-xs text-white">
          from{' '}
          {image.shared_by.username
            ? `@${image.shared_by.username}`
            : 'a circle member'}
        </span>
      )}
    </button>
  );
}
