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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { MoreVertical } from "lucide-react";

type ClientActionsProps = {
  client: {
    allowed_scopes: string[];
    id: string;
    name: string;
    redirect_uri: string;
    revoked_at?: string | null;
  };
  updateClientAction: (formData: FormData) => void | Promise<void>;
  revokeClientAction: (formData: FormData) => void | Promise<void>;
  deleteClientAction: (formData: FormData) => void | Promise<void>;
};

export function ClientActions({
  client,
  updateClientAction,
  revokeClientAction,
  deleteClientAction,
}: ClientActionsProps) {
  const [editOpen, setEditOpen] = React.useState(false);
  const [revokeOpen, setRevokeOpen] = React.useState(false);
  const [deleteOpen, setDeleteOpen] = React.useState(false);
  const isRevoked = Boolean(client.revoked_at);

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon">
            <MoreVertical className="h-4 w-4" />
            <span className="sr-only">Open client menu</span>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem
            disabled={isRevoked}
            onSelect={(event) => {
              event.preventDefault();
              if (!isRevoked) {
                setEditOpen(true);
              }
            }}
          >
            Edit client
          </DropdownMenuItem>
          <DropdownMenuItem
            disabled={isRevoked}
            onSelect={(event) => {
              event.preventDefault();
              if (!isRevoked) {
                setRevokeOpen(true);
              }
            }}
          >
            Revoke client
          </DropdownMenuItem>
          <DropdownMenuItem
            onSelect={(event) => {
              event.preventDefault();
              setDeleteOpen(true);
            }}
          >
            Delete client
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit client</DialogTitle>
            <DialogDescription>
              Update the details for {client.name}.
            </DialogDescription>
          </DialogHeader>

          <form action={updateClientAction} className="space-y-4">
            <input type="hidden" name="client_id" value={client.id} />

            <div className="space-y-2">
              <Label htmlFor={`client-name-${client.id}`}>Name</Label>
              <Input
                id={`client-name-${client.id}`}
                name="name"
                defaultValue={client.name}
                required
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor={`client-redirect-uri-${client.id}`}>Redirect URI</Label>
              <Input
                id={`client-redirect-uri-${client.id}`}
                name="redirect_uri"
                defaultValue={client.redirect_uri}
                type="url"
                required
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor={`client-scopes-${client.id}`}>Allowed scopes</Label>
              <Input
                id={`client-scopes-${client.id}`}
                name="allowed_scopes"
                defaultValue={client.allowed_scopes.join(" ")}
                required
              />
            </div>

            <DialogFooter>
              <Button type="submit" disabled={isRevoked}>
                Save changes
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={revokeOpen} onOpenChange={setRevokeOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Revoke client</DialogTitle>
            <DialogDescription>
              Revoke {client.name} so it can no longer be used for sign-in.
            </DialogDescription>
          </DialogHeader>

          <form action={revokeClientAction}>
            <input type="hidden" name="client_id" value={client.id} />
            <DialogFooter>
              <Button type="submit" variant="destructive" disabled={isRevoked}>
                Revoke client
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete client</DialogTitle>
            <DialogDescription>
              Permanently delete {client.name}. This action cannot be undone.
            </DialogDescription>
          </DialogHeader>

          <form action={deleteClientAction}>
            <input type="hidden" name="client_id" value={client.id} />
            <DialogFooter>
              <Button type="submit" variant="destructive">
                Delete client
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </>
  );
}
