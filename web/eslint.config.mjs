// Galileo OS — web/ ESLint flat config.
//
// Bridges eslint-config-next (legacy preset) into ESLint 9 flat config
// via @eslint/eslintrc's FlatCompat. The Stage 0 stub uses Next.js
// Core Web Vitals + TypeScript presets; rules will be tightened as
// the admin app grows.

import { FlatCompat } from "@eslint/eslintrc";
import { fileURLToPath } from "node:url";
import { dirname } from "node:path";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const compat = new FlatCompat({ baseDirectory: __dirname });

export default [
  {
    ignores: [".next/**", "node_modules/**", "out/**"],
  },
  ...compat.extends("next/core-web-vitals", "next/typescript"),
];
