import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import {
  getGetCirclesIdMembersQueryKey,
  usePatchCirclesIdMembersUserId,
} from '@/lib/api/generated/circles/circles';
import { queryClient } from '@/lib/query-client';

interface CirclePermissionToggleProps {
  circleId: string;
  userId: string;
  label: string;
  field: 'can_invite' | 'can_capture';
  canInvite: boolean;
  canCapture: boolean;
}

export function CirclePermissionToggle({
  circleId,
  userId,
  label,
  field,
  canInvite,
  canCapture,
}: CirclePermissionToggleProps) {
  const updatePermissions = usePatchCirclesIdMembersUserId();
  const checked = field === 'can_invite' ? canInvite : canCapture;

  const handleChange = async (next: boolean) => {
    await updatePermissions.mutateAsync({
      id: circleId,
      userId,
      data: {
        can_invite: field === 'can_invite' ? next : canInvite,
        can_capture: field === 'can_capture' ? next : canCapture,
      },
    });
    await queryClient.invalidateQueries({
      queryKey: getGetCirclesIdMembersQueryKey(circleId),
    });
  };

  const inputId = `${field}-${userId}`;

  return (
    <div className="flex items-center gap-2">
      <Checkbox
        id={inputId}
        checked={checked}
        disabled={updatePermissions.isPending}
        onCheckedChange={handleChange}
      />
      <Label htmlFor={inputId} className="text-sm text-muted-foreground">
        {label}
      </Label>
    </div>
  );
}
