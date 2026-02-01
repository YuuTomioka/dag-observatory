export default {
  dagCore: {
    input: "./openapi.yaml",
    output: {
      target: "./gen/ts/index.ts",
      schemas: "./gen/ts/model",
      client: "react-query",
      override: {
        mutator: {
          path: "./gen/ts/mutator.ts",
          name: "mutator",
        },
      },
    },
  },
};
