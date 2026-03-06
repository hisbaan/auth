import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { verifyEmailAction } from "@/lib/actions";

type VerifyEmailProps = {
  searchParams: Promise<{ token?: string }>;
};

export default async function VerifyEmailPage({
  searchParams,
}: VerifyEmailProps) {
  const params = await searchParams;

  return (
    <main className="mx-auto flex w-full max-w-6xl justify-center px-4 py-10 sm:px-6">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Verify email</CardTitle>
          <CardDescription>
            Submit your verification token to confirm ownership.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <form action={verifyEmailAction} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="token">Verification token</Label>
              <Input
                id="token"
                name="token"
                defaultValue={params.token ?? ""}
                required
              />
            </div>

            <Button className="w-full" type="submit">
              Verify email
            </Button>
          </form>
        </CardContent>
      </Card>
    </main>
  );
}
