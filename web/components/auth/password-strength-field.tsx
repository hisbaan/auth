"use client";

import { useId, useState } from "react";

import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

const MIN_PASSWORD_LENGTH = 8;
const MAX_PASSWORD_LENGTH = 1024;

const PASSWORD_RULES = [
  {
    label: `At least ${MIN_PASSWORD_LENGTH} characters`,
    test: (value: string) => value.length >= MIN_PASSWORD_LENGTH,
  },
  {
    label: "At least one uppercase letter",
    test: (value: string) => /\p{Lu}/u.test(value),
  },
  {
    label: "At least one lowercase letter",
    test: (value: string) => /\p{Ll}/u.test(value),
  },
  {
    label: "At least one number",
    test: (value: string) => /\p{Nd}/u.test(value),
  },
  {
    label: "At least one symbol or punctuation mark",
    test: (value: string) => /[\p{P}\p{S}]/u.test(value),
  },
];

const PASSWORD_REQUIREMENTS = `Use uppercase, lowercase, a number, and a special character.`;

type PasswordStrengthFieldProps = {
  id: string;
  name: string;
  label: string;
  autoComplete?: string;
  className?: string;
};

export function PasswordStrengthField({
  id,
  name,
  label,
  autoComplete,
  className,
}: PasswordStrengthFieldProps) {
  const descriptionId = useId();
  const [password, setPassword] = useState("");
  const [failedRule] = PASSWORD_RULES.filter((rule) => !rule.test(password));
  const isInvalid = password.length > 0 && failedRule != null;

  return (
    <div className={cn("space-y-2", className)}>
      <Label htmlFor={id}>{label}</Label>
      <Input
        id={id}
        name={name}
        type="password"
        required
        minLength={MIN_PASSWORD_LENGTH}
        maxLength={MAX_PASSWORD_LENGTH}
        autoComplete={autoComplete}
        aria-describedby={descriptionId}
        aria-invalid={isInvalid}
        onChange={(event) => {
          const nextPassword = event.currentTarget.value;
          const nextFailedRules = PASSWORD_RULES.filter((rule) => !rule.test(nextPassword));

          event.currentTarget.setCustomValidity(
            nextFailedRules.length === 0 ? "" : "Password does not meet the strength requirements",
          );
          setPassword(nextPassword);
        }}
      />
      <p
        id={descriptionId}
        className={cn("text-xs text-muted-foreground", isInvalid && "text-destructive")}
      >
        {password.length === 0
          ? PASSWORD_REQUIREMENTS
          : failedRule
            ? failedRule.label
            : "Password meets strength requirements"}
      </p>
    </div>
  );
}
