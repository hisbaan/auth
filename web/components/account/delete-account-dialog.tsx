"use client";

import * as React from "react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

type DeleteAccountDialogProps = {
  email: string;
  deleteAccountAction: (formData: FormData) => void | Promise<void>;
};

export function DeleteAccountDialog({ email, deleteAccountAction }: DeleteAccountDialogProps) {
  const [confirmation, setConfirmation] = React.useState("");
  const isConfirmed = confirmation === email;

  return (
    <Dialog onOpenChange={(open) => !open && setConfirmation("")}>
      <DialogTrigger asChild>
        <Button type="button" variant="destructive">
          Delete account
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete account</DialogTitle>
          <DialogDescription>
            This permanently deletes your account. Type your email address to confirm.
          </DialogDescription>
        </DialogHeader>

        <form action={deleteAccountAction} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="delete-account-email">Email address</Label>
            <Input
              id="delete-account-email"
              name="email"
              type="email"
              autoComplete="email"
              value={confirmation}
              onChange={(event) => setConfirmation(event.target.value)}
              placeholder={email}
              required
            />
          </div>

          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="outline">
                Cancel
              </Button>
            </DialogClose>
            <Button type="submit" variant="destructive" disabled={!isConfirmed}>
              Delete account
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
