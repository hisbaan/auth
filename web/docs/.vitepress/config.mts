import { defineConfig } from "vitepress";

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "auth docs",
  description: "auth docs",
  outDir: "../public/docs",
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config

    sidebar: [
      {
        text: "OpenID Connect",
        items: [
          { text: "Overview", link: "/oidc/" },
          {
            text: "Authorization Code Flow",
            link: "/oidc/authorization-code-flow",
          },
          { text: "Tokens And UserInfo", link: "/oidc/tokens-and-userinfo" },
          { text: "Token Validation", link: "/oidc/token-validation" },
          { text: "Errors", link: "/oidc/errors" },
          { text: "Client Checklist", link: "/oidc/client-checklist" },
        ],
      },
    ],

    socialLinks: [
      { icon: "github", link: "https://github.com/hisbaan/auth" },
      { icon: "googlehome", link: "https://auth.hisbaan.com" },
    ],
  },
});
