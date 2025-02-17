import { defineConfig } from "vitepress";

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Ika Gateway",
  description:
    "High-performance API Gateway with Go plugins. Zero bloat, sub-millisecond routing, and enterprise-grade reliability for modern microservices.",
  head: [["link", { rel: "icon", href: "/favicon.png" }]],
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
            { text: "Showcase", link: "/guide/showcase" },
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
          items: [{ text: "Overview", link: "/roadmap" }],
        },
      ],
    },

    search: {
      provider: "local",
    },

    socialLinks: [{ icon: "github", link: "https://github.com/alx99/ika" }],

    siteTitle: "Ika Gateway",
    logo: { light: "/ika.webp", dark: "/ika.webp" },
  },
  sitemap: {
    hostname: "https://ika.dozy.dev",
  },
});
