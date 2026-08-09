import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemSeparator,
} from '@/components/ui/item';
import { CircleUserChip } from './circle-user-chip';
import { CirclePermissionToggle } from './circle-permission-toggle';
import {
  getGetCirclesIdMembersQueryKey,
  useDeleteCirclesIdMembersUserId,
  useGetCirclesIdMembers,
  usePatchCirclesIdMembersUserIdRole,
} from '@/lib/api/generated/circles/circles';
import { queryClient } from '@/lib/query-client';
import { MoreHorizontalIcon } from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';

interface CircleMemberListProps {
  circleId: string;
  viewerUserId: string;
  viewerIsAdmin: boolean;
}

export function CircleMemberList({
  circleId,
  viewerUserId,
  viewerIsAdmin,
}: CircleMemberListProps) {
  const membersQuery = useGetCirclesIdMembers(circleId);
  const removeMember = useDeleteCirclesIdMembersUserId();
  const updateRole = usePatchCirclesIdMembersUserIdRole();

  const invalidate = () =>
    queryClient.invalidateQueries({
      queryKey: getGetCirclesIdMembersQueryKey(circleId),
    });

  const members = membersQuery.data?.members ?? [];

  if (membersQuery.isPending) {
    return <p className="text-muted-foreground">Loading members…</p>;
  }

  if (members.length === 0) {
    return <p className="text-muted-foreground">No members yet.</p>;
  }

  return (
    <ItemGroup>
      {members.map((member, index) => {
        const canManage = viewerIsAdmin && member.user_id !== viewerUserId;

        return (
          <div key={member.user_id}>
            <Item size="sm">
              <ItemContent className="flex-row flex-wrap items-center gap-4">
                <CircleUserChip userId={member.user_id!} />
                <ItemDescription>
                  {member.role === 'admin' ? 'Admin' : 'Member'}
                </ItemDescription>
                {canManage && (
                  <div className="flex items-center gap-4">
                    <CirclePermissionToggle
                      circleId={circleId}
                      userId={member.user_id!}
                      field="can_invite"
                      label="Can invite"
                      canInvite={member.can_invite ?? false}
                      canCapture={member.can_capture ?? false}
                    />
                    <CirclePermissionToggle
                      circleId={circleId}
                      userId={member.user_id!}
                      field="can_capture"
                      label="Can capture"
                      canInvite={member.can_invite ?? false}
                      canCapture={member.can_capture ?? false}
                    />
                  </div>
                )}
              </ItemContent>
              {canManage && (
                <ItemActions>
                  <DropdownMenu>
                    <DropdownMenuTrigger
                      render={<Button variant="ghost" size="icon" />}
                    >
                      <HugeiconsIcon
                        icon={MoreHorizontalIcon}
                        strokeWidth={2}
                      />
                    </DropdownMenuTrigger>
                    <DropdownMenuContent>
                      <DropdownMenuGroup>
                        <DropdownMenuItem
                          onClick={async () => {
                            await updateRole.mutateAsync({
                              id: circleId,
                              userId: member.user_id!,
                              data: {
                                role:
                                  member.role === 'admin' ? 'member' : 'admin',
                              },
                            });
                            await invalidate();
                          }}
                        >
                          {member.role === 'admin'
                            ? 'Demote to member'
                            : 'Promote to admin'}
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          variant="destructive"
                          onClick={async () => {
                            await removeMember.mutateAsync({
                              id: circleId,
                              userId: member.user_id!,
                            });
                            await invalidate();
                          }}
                        >
                          Remove from circle
                        </DropdownMenuItem>
                      </DropdownMenuGroup>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </ItemActions>
              )}
            </Item>
            {index !== members.length - 1 && <ItemSeparator />}
          </div>
        );
      })}
    </ItemGroup>
  );
}
