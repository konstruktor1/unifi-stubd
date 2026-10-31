import js from "@eslint/js";

export default [
  js.configs.recommended,
  {
    languageOptions: {
      globals: {
        AbortSignal: "readonly",
        fetch: "readonly",
        process: "readonly",
        Response: "readonly",
        URL: "readonly",
      },
    },
  },
];
