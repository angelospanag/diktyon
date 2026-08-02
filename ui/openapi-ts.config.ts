import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  input: "../api/openapi.yaml",
  output: {
    path: "client",
    postProcess: ["biome:format"],
  },
  plugins: ["@hey-api/client-next", "@hey-api/typescript", "@hey-api/sdk"],
});
