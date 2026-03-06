"use client";

import { redirect } from "next/navigation";

import { CopyableCode } from "@/components/ui/copyable-code";
import { TableCell, TableRow } from "@/components/ui/table";
import { UserRoleActions } from "@/components/admin/user-role-actions";
import type { User } from "@/lib/types";
import { addUserRoleAction, removeUserRoleAction } from "@/lib/actions";

type AdminUserRowProps = {
  user: User;
  roles: string[];
};

export function AdminUserRow({
  user,
  roles,
}: AdminUserRowProps) {
  const userRoles = user.roles ?? [];

  return (
    <TableRow
      className={"cursor-pointer"}
      onClick={() => {
        redirect(`/admin/users/${user.id}`);
      }}
      role={"link"}
      tabIndex={0}
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") {
          redirect(`/admin/users/${user.id}`);
        }
      }}
    >
      <TableCell className="font-mono">{user.username}</TableCell>
      <TableCell className="font-mono">{user.email}</TableCell>
      <TableCell onClick={(event) => event.stopPropagation()}>
        <CopyableCode value={user.id} />
      </TableCell>
      <TableCell className="font-mono">{userRoles.join(", ")}</TableCell>
      <TableCell className="font-mono">
        {user.email_verified ? "true" : "false"}
      </TableCell>
      <TableCell onClick={(event) => event.stopPropagation()}>
        <UserRoleActions
          user={user}
          roles={roles}
          addUserRoleAction={addUserRoleAction}
          removeUserRoleAction={removeUserRoleAction}
          viewHref={`/admin/users/${user.id}`}
        />
      </TableCell>
    </TableRow>
  );
}
