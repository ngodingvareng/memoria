import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { UserRestrictionList } from './user-restriction-list';

export function BlockedMutedUsersList() {
  return (
    <Tabs defaultValue="blocked">
      <TabsList>
        <TabsTrigger value="blocked">Blocked</TabsTrigger>
        <TabsTrigger value="muted">Muted</TabsTrigger>
      </TabsList>
      <TabsContent value="blocked">
        <UserRestrictionList kind="blocked" />
      </TabsContent>
      <TabsContent value="muted">
        <UserRestrictionList kind="muted" />
      </TabsContent>
    </Tabs>
  );
}
