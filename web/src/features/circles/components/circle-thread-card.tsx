import { Item, ItemContent, ItemHeader, ItemTitle } from '@/components/ui/item';
import type { GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadResponse } from '@/lib/api/generated/models';
import { Link } from '@tanstack/react-router';

interface CircleThreadCardProps {
  thread: GithubComNgodingvarengMemoriaInternalDeliveryRestDtoThreadResponse;
}

export function CircleThreadCard({ thread }: CircleThreadCardProps) {
  return (
    <Item
      variant="outline"
      render={<Link to="/thread/$id" params={{ id: thread.id! }} />}
    >
      <ItemHeader>
        <div
          className="aspect-video w-full rounded-sm"
          style={{ backgroundColor: thread.color_hex ?? '#374151' }}
        />
      </ItemHeader>
      <ItemContent>
        <ItemTitle>{thread.name}</ItemTitle>
      </ItemContent>
    </Item>
  );
}
