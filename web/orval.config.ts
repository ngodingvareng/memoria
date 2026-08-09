import { defineConfig } from 'orval';

export default defineConfig({
  memoria: {
    input: {
      target: '../api/docs/swagger.json',
      override: {
        transformer: './src/lib/api/transformer.ts',
      },
    },
    output: {
      mode: 'tags-split',
      target: './src/lib/api/generated',
      schemas: './src/lib/api/generated/models',
      client: 'react-query',
      httpClient: 'fetch',
      clean: true,
      formatter: 'prettier',
      indexFiles: true,
      override: {
        mutator: {
          path: './src/lib/api/mutator.ts',
          name: 'apiMutator',
        },
        fetch: {
          // apiMutator already unwraps the WebResponse envelope down to
          // `.data` and throws ApiError on failure, so generated hooks
          // should return that plain success type directly instead of
          // orval's default { data, status, headers } wrapper.
          includeHttpResponseReturnType: false,
        },
      },
    },
  },
});
