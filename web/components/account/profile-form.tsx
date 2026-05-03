"use client";

import * as React from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

type ProfileFormProps = {
  email: string;
  roles: string[];
  updateProfileAction: (formData: FormData) => void | Promise<void>;
  username: string;
};

export function ProfileForm({ email, roles, updateProfileAction, username }: ProfileFormProps) {
  const [currentUsername, setCurrentUsername] = React.useState(username);
  const [currentEmail, setCurrentEmail] = React.useState(email);
  const isDirty = currentUsername !== username || currentEmail !== email;

  return (
    <form action={updateProfileAction} className="space-y-4">
      <div className="flex flex-col gap-2">
        <Label htmlFor="username">Username</Label>
        <Input
          id="username"
          name="username"
          value={currentUsername}
          onChange={(event) => setCurrentUsername(event.target.value)}
          required
          maxLength={64}
        />
      </div>
      <div className="flex flex-col gap-2">
        <Label htmlFor="email">Email</Label>
        <Input
          id="email"
          name="email"
          type="email"
          value={currentEmail}
          onChange={(event) => setCurrentEmail(event.target.value)}
          required
          maxLength={254}
        />
      </div>
      <div className="flex flex-wrap gap-2">
        {roles.map((role) => (
          <Badge key={role} variant="secondary">
            {role}
          </Badge>
        ))}
      </div>
      <Button type="submit" disabled={!isDirty}>Save Changes</Button>
    </form>
  );
}
