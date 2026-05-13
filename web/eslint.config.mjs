// Galileo OS — web/ ESLint flat config (native, no FlatCompat).
//
// Stage 0 uses @eslint/js recommended + typescript-eslint strict for
// the two stub .tsx files. eslint-config-next is intentionally NOT
// extended: bridging it through @eslint/eslintrc's FlatCompat surfaces
// circular-JSON errors (the plugin self-references break serialization),
// and the two stubs have no Next.js-specific patterns to lint anyway.
// When Stage 1 ships real Next.js code, eslint-config-next will likely
// have native flat-config support and is added back in a single line.
// See docs/solutions/SOLUTION_CI_EXPANSION_FINDINGS.md.

import js from "@eslint/js";
import tseslint from "typescript-eslint";

export default tseslint.config(
  {
    ignores: [".next/**", "node_modules/**", "out/**"],
  },
  js.configs.recommended,
  ...tseslint.configs.strict,
  ...tseslint.configs.stylistic,
);
