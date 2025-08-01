module.exports = {
  root: true,
  parser: "@typescript-eslint/parser",
  plugins: ["@typescript-eslint", "prettier", "import"],
  parserOptions: {
    project: "./tsconfig.json",
    createDefaultProgram: true,
  },
  env: {
    // 전역객체를 eslint가 인식하는 구간
    browser: true, // document나 window 인식되게 함
    node: true,
    es6: true,
  },
  ignorePatterns: ["node_modules/"], // eslint 미적용될 폴더나 파일 명시
  extends: [
    "airbnb",
    "airbnb-typescript",
    "airbnb/hooks",
    "next/core-web-vitals",
    "plugin:@typescript-eslint/recommended",
    "plugin:prettier/recommended", // eslint의 포매팅을 prettier로 사용
    "prettier", // eslint-config-prettier prettier와 중복된 eslint 규칙 제거
    "eslint-config-prettier",
    "plugin:import/recommended",
    "plugin:import/typescript",
    "eslint:recommended",
    "plugin:react/recommended",
    "plugin:react-hooks/recommended",
  ],
  settings: {
    "import/resolver": {
      typescript: {},
    },
    "import/parsers": { "@typescript-eslint/parser": [".ts"] },
  },
  rules: {
    "@typescript-eslint/no-explicit-any": "off",
    "react/react-in-jsx-scope": "off", // react 17부턴 import 안해도돼서 기능 끔
    // 경고표시, 파일 확장자를 .ts나 .tsx 모두 허용함
    "react/jsx-filename-extension": ["warn", { extensions: [".ts", ".tsx"] }],
    "react/function-component-definition": [
      2,
      { namedComponents: ["arrow-function", "function-declaration"] },
    ],
    "sort-imports": [
      "error",
      {
        ignoreCase: false,
        ignoreDeclarationSort: false,
        ignoreMemberSort: false,
        memberSyntaxSortOrder: ["none", "all", "multiple", "single"],
        allowSeparatedGroups: false,
      },
    ],
    "import/order": [
      "error",
      {
        groups: [
          ["builtin", "external"],
          "internal",
          ["parent", "sibling"],
          "index",
          "object",
        ],
        pathGroups: [
          {
            pattern: "~/**",
            group: "external",
            position: "before",
          },
          { pattern: "@*", group: "external", position: "after" },
          { pattern: "@*/**", group: "external", position: "after" },
        ],
        pathGroupsExcludedImportTypes: ["react"],
        "newlines-between": "always",
        alphabetize: {
          order: "asc",
          caseInsensitive: true,
        },
      },
    ],
  },
};
