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
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

type CreateClientDialogProps = {
  createClientAction: (formData: FormData) => void | Promise<void>;
};

export function CreateClientDialog({
  createClientAction,
}: CreateClientDialogProps) {
  const [open, setOpen] = React.useState(false);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>+ New</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create client</DialogTitle>
          <DialogDescription>
            Register a public OIDC client for your application.
          </DialogDescription>
        </DialogHeader>

        <form action={createClientAction} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              name="name"
              placeholder="Acme Dashboard"
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="redirect_uri">Redirect URI</Label>
            <Input
              id="redirect_uri"
              name="redirect_uri"
              placeholder="https://app.example.com/auth/callback"
              type="url"
              required
            />
            <p className="text-xs text-muted-foreground">
              Enter the exact callback URL your app uses after sign-in.
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="allowed_scopes">Allowed scopes</Label>
            <Input
              id="allowed_scopes"
              name="allowed_scopes"
              placeholder="openid profile email"
              required
            />
            <p className="text-xs text-muted-foreground">
              Use space-separated scopes. `openid` is added automatically if omitted.
            </p>
          </div>

          <DialogFooter>
            <Button type="submit">Create client</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
