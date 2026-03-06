import { Disc, Github, Gitlab, Mail } from "lucide-react";

import { Button } from "@/components/ui/button";

const providers = [
  { label: "Google", icon: Mail },
  { label: "Discord", icon: Disc },
  { label: "GitHub", icon: Github },
  { label: "GitLab", icon: Gitlab },
];

export function SocialButtons() {
  return (
    <div className="space-y-3">
      {providers.map((provider) => {
        const Icon = provider.icon;

        return (
          <Button key={provider.label} type="button" variant="outline" className="w-full justify-start gap-2" disabled>
            <Icon className="h-4 w-4" />
            Continue with {provider.label}
            <span className="ml-auto text-xs text-muted-foreground">Soon</span>
          </Button>
        );
      })}
    </div>
  );
}
