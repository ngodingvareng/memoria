import { Button } from '@/components/ui/button';
import { ButtonGroup } from '@/components/ui/button-group';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuPortal,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Item, ItemContent } from '@/components/ui/item';
import {
  ArrowUp02Icon,
  Attachment01Icon,
  FileEmpty02Icon,
  Heading01Icon,
  Heading02Icon,
  Heading03Icon,
  HeadingIcon,
  MultiplicationSignIcon,
  PlusSignIcon,
} from '@hugeicons/core-free-icons';
import { HugeiconsIcon } from '@hugeicons/react';
import {
  headingsPlugin,
  linkPlugin,
  markdownShortcutPlugin,
  MDXEditor,
} from '@mdxeditor/editor';
import '@mdxeditor/editor/style.css';

interface CaptureInputProps {
  onOpenTimeDialog: () => void;
  onPublish: () => void;
}

export function CaptureInput({
  onOpenTimeDialog,
  onPublish,
}: CaptureInputProps) {
  return (
    <div className="sticky bottom-0 pb-6 left-0 z-30 bg-linear-to-t from-background pt-20 from-60% to-transparent w-full">
      <Item variant="outline" className="mx-auto shadow-sm bg-card max-w-5xl">
        <ItemContent className="flex flex-col min-h-20 max-h-[calc(100vh-10rem)]">
          <div className="grow overflow-y-auto  ">
            <MDXEditor
              markdown=""
              placeholder="Woylah cikk, ketik sini..."
              contentEditableClassName="typeset! max-w-none!"
              plugins={[
                headingsPlugin(),
                linkPlugin(),
                markdownShortcutPlugin(),
              ]}
            />
          </div>
          <div className="flex justify-between gap-1">
            <div>
              <DropdownMenu>
                <DropdownMenuTrigger
                  render={<Button variant="secondary" size="sm" />}
                >
                  <HugeiconsIcon strokeWidth={2.5} icon={PlusSignIcon} />
                  Add files and more...
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <DropdownMenuSub>
                    <DropdownMenuSubTrigger>
                      <HugeiconsIcon strokeWidth={2} icon={HeadingIcon} />
                      Heading
                    </DropdownMenuSubTrigger>
                    <DropdownMenuPortal>
                      <DropdownMenuSubContent>
                        <DropdownMenuItem>
                          <HugeiconsIcon strokeWidth={2} icon={Heading01Icon} />
                          H1
                        </DropdownMenuItem>
                        <DropdownMenuItem>
                          <HugeiconsIcon strokeWidth={2} icon={Heading02Icon} />
                          H2
                        </DropdownMenuItem>
                        <DropdownMenuItem>
                          <HugeiconsIcon strokeWidth={2} icon={Heading03Icon} />
                          H3
                        </DropdownMenuItem>
                      </DropdownMenuSubContent>
                    </DropdownMenuPortal>
                  </DropdownMenuSub>
                  <DropdownMenuItem>
                    <HugeiconsIcon strokeWidth={2} icon={FileEmpty02Icon} />
                    Photos & files
                  </DropdownMenuItem>
                  <DropdownMenuItem>
                    <HugeiconsIcon strokeWidth={2} icon={Attachment01Icon} />
                    Links
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>

            <div className="flex items-center gap-2">
              <Button variant="ghost" size="sm" onClick={onPublish}>
                <HugeiconsIcon icon={MultiplicationSignIcon} />
                Publish
              </Button>
              <ButtonGroup>
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={onOpenTimeDialog}
                >
                  10/02/2033 10:00
                </Button>
                <Button size="icon-sm">
                  <HugeiconsIcon strokeWidth={2.5} icon={ArrowUp02Icon} />
                </Button>
              </ButtonGroup>
            </div>
          </div>
        </ItemContent>
      </Item>
    </div>
  );
}
