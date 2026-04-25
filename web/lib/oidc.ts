export type OIDCScopeDetail = {
  title: string;
  description: string;
};

const OIDC_SCOPE_DETAILS: Record<string, OIDCScopeDetail> = {
  openid: {
    title: "Confirm your identity",
    description: "Lets the app verify who you are using your auth account.",
  },
  profile: {
    title: "View your profile",
    description: "Shares your username with the app.",
  },
  email: {
    title: "View your email",
    description: "Shares your email address and verification status.",
  },
};

export function getOIDCScopeDetail(scope: string): OIDCScopeDetail {
  return (
    OIDC_SCOPE_DETAILS[scope] ?? {
      title: scope,
      description: "Allows the app to use this scope.",
    }
  );
}
