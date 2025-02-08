import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { federation } from "@module-federation/vite";

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    federation({
      name: "tbz-baumkataster",    // hier slug als name verwenden
      filename: "plugin.js",       // hier immer plugin.js
      exposes: {
        "./app": "./src/App.tsx",  // hier ./app
      },
      shared: ["react", "react-dom", "@green-ecolution/plugin-interface"],
    }),
  ],
  base: "",
  build: {
    target: "esnext",
  },
});
