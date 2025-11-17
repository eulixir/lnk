import { defineConfig } from "orval";

export default defineConfig({
  "lnk-file": {
    input: {
      target: "http://localhost:8080/swagger/doc.json",
      validation: false,
    },
    output: {
      target: "./src/api/lnk.ts",
      client: "axios",
      override: {
        mutator: {
          path: "./src/api/axios-instance.ts",
          name: "customInstance",
          default: false,
        },
      },
    },
  },
});
