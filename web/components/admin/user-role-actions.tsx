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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { MoreVertical } from "lucide-react";
import Link from "next/link";

type User = {
  id: string;
  username: string;
  roles?: string[];
};

type UserRoleActionsProps = {
  user: User;
  roles: string[];
  addUserRoleAction: (formData: FormData) => void | Promise<void>;
  removeUserRoleAction: (formData: FormData) => void | Promise<void>;
  viewHref?: string;
};

export function UserRoleActions({
  user,
  roles,
  addUserRoleAction,
  removeUserRoleAction,
  viewHref,
}: UserRoleActionsProps) {
  const [addOpen, setAddOpen] = React.useState(false);
  const [removeOpen, setRemoveOpen] = React.useState(false);

  const userRoles = user.roles ?? [];
  const addableRoles = roles.filter((role) => !userRoles.includes(role));
  const removableRoles = userRoles;

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon">
            <MoreVertical className="h-4 w-4" />
            <span className="sr-only">Open menu</span>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          {viewHref ? (
            <DropdownMenuItem asChild>
              <Link href={viewHref}>View user</Link>
            </DropdownMenuItem>
          ) : null}
          <DropdownMenuItem
            onSelect={(event) => {
              event.preventDefault();
              setAddOpen(true);
            }}
          >
            Add role
          </DropdownMenuItem>
          <DropdownMenuItem
            onSelect={(event) => {
              event.preventDefault();
              setRemoveOpen(true);
            }}
          >
            Remove role
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <Dialog open={addOpen} onOpenChange={setAddOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add role</DialogTitle>
            <DialogDescription>
              Assign a role to {user.username}.
            </DialogDescription>
          </DialogHeader>

          <form action={addUserRoleAction}>
            <input type="hidden" name="user_id" value={user.id} />
            <div className="space-y-4">
              <Select
                name="role"
                required
                disabled={addableRoles.length === 0}
              >
                <SelectTrigger className="w-full">
                  <SelectValue
                    placeholder={
                      addableRoles.length === 0
                        ? "No available roles"
                        : "Select role to add"
                    }
                  />
                </SelectTrigger>
                <SelectContent>
                  {addableRoles.map((role) => (
                    <SelectItem key={role} value={role}>
                      {role}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <DialogFooter>
                <Button type="submit" disabled={addableRoles.length === 0}>
                  Add role
                </Button>
              </DialogFooter>
            </div>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={removeOpen} onOpenChange={setRemoveOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Remove role</DialogTitle>
            <DialogDescription>
              Remove a role from {user.username}.
            </DialogDescription>
          </DialogHeader>

          <form action={removeUserRoleAction}>
            <input type="hidden" name="user_id" value={user.id} />
            <div className="space-y-4">
              <Select
                name="role"
                required
                disabled={removableRoles.length === 0}
              >
                <SelectTrigger className="w-full">
                  <SelectValue
                    placeholder={
                      removableRoles.length === 0
                        ? "No roles to remove"
                        : "Select role to remove"
                    }
                  />
                </SelectTrigger>
                <SelectContent>
                  {removableRoles.map((role) => (
                    <SelectItem key={role} value={role}>
                      {role}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <DialogFooter>
                <Button
                  type="submit"
                  variant="destructive"
                  disabled={removableRoles.length === 0}
                >
                  Remove role
                </Button>
              </DialogFooter>
            </div>
          </form>
        </DialogContent>
      </Dialog>
    </>
  );
}
