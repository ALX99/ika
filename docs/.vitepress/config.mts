import { defineConfig } from "vitepress";

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Ika Gateway",
  description:
    "Documentation for the worlds-most minimal, flexible, and performant API Gateway, Ika",
  ignoreDeadLinks: [/^https?:\/\/localhost/, (url) => url.includes("TODO")],
  cleanUrls: true,
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: "Guide", link: "/guide/why-ika" },
      { text: "Plugins", link: "/plugins/" },
      { text: "Roadmap", link: "/roadmap" },
    ],

    sidebar: {
      "/guide/": [
        {
          text: "Introduction",
          collapsed: false,
          items: [
            { text: "Why Ika", link: "/guide/why-ika" },
            { text: "Motivation", link: "/guide/motivation" },
            { text: "Getting Started", link: "/guide/getting-started" },
          ],
        },
      ],
      "/plugins/": [
        {
          text: "Core Plugins",
          items: [
            { text: "Overview", link: "/plugins/" },
            { text: "Access Log", link: "/plugins/access-log" },
            { text: "Basic Auth", link: "/plugins/basic-auth" },
            { text: "Request ID", link: "/plugins/request-id" },
            { text: "Request Modifier", link: "/plugins/request-modifier" },
            { text: "Fail2Ban", link: "/plugins/fail2ban" },
          ],
        },
      ],
      "/roadmap": [
        {
          text: "Project Roadmap",
          items: [
            { text: "Overview", link: "/roadmap" },
          ],
        },
      ],
    },

    socialLinks: [{ icon: "github", link: "https://github.com/alx99/ika" }],
  },
  sitemap: {
    hostname: "https://ika.dozy.dev",
  },
});
