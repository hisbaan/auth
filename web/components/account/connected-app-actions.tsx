"use client";

import * as React from "react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { ClientAuthorization } from "@/lib/types";
import { MoreVertical } from "lucide-react";

type ConnectedAppActionsProps = {
  authorization: ClientAuthorization;
  revokeAuthorizationAction: (formData: FormData) => void | Promise<void>;
};

export function ConnectedAppActions({
  authorization,
  revokeAuthorizationAction,
}: ConnectedAppActionsProps) {
  const [disconnectOpen, setDisconnectOpen] = React.useState(false);

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon">
            <MoreVertical className="h-4 w-4" />
            <span className="sr-only">Open connected app menu</span>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem
            onSelect={(event) => {
              event.preventDefault();
              setDisconnectOpen(true);
            }}
          >
            Disconnect app
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <Dialog open={disconnectOpen} onOpenChange={setDisconnectOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Disconnect app</DialogTitle>
            <DialogDescription>
              Disconnect {authorization.name} from your account. You can grant
              access again the next time you sign in with this app.
            </DialogDescription>
          </DialogHeader>

          <form action={revokeAuthorizationAction}>
            <input type="hidden" name="client_id" value={authorization.client_id} />
            <DialogFooter>
              <Button type="submit" variant="destructive">
                Disconnect app
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </>
  );
}
