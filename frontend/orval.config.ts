import { defineConfig } from "orval";

export default defineConfig({
  "lnk-file": {
    input: {
      target: "http://localhost:8080/swagger/doc.json",
      validation: false,
    },
    output: {
      target: "./src/api/lnk.ts",
      client: "fetch",
      override: {
        mutator: {
          path: "./src/api/undici-instance.ts",
          name: "customInstance",
          default: false,
        },
      },
    },
  },
});
