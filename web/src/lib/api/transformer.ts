// Runs on the raw Swagger 2.0 doc before Orval upgrades it to 3.1 and
// generates types/hooks from it. Two fixes are needed here because they'd
// otherwise leak straight into the generated code:
//
// 1. Every response in api/'s swagger.json is wrapped in a generic
//    dto.WebResponse[T] envelope ({code, message, data, errors}), so
//    without this, every generated type would be the wrapper shape
//    (WebResponse-XxxResponse) instead of T itself. apiFetch (see
//    src/lib/api-client.ts) already unwraps `.data` at runtime — this
//    collapses each WebResponse-* definition down to its `data` schema so
//    generated types match what apiFetch actually returns.
// 2. main.go's `@BasePath /api/v1` swag annotation is stale — the live
//    Fiber app mounts routes with no /api/v1 prefix at all (no
//    app.Group("/api/v1") exists) — so basePath is blanked out to keep
//    generated call URLs matching the real server.

interface SwaggerSchema {
  properties?: Record<string, unknown>;
  [key: string]: unknown;
}

interface SwaggerSpec {
  basePath?: string;
  definitions?: Record<string, SwaggerSchema>;
  [key: string]: unknown;
}

export default function transformer(spec: SwaggerSpec): SwaggerSpec {
  if (spec.definitions) {
    for (const [name, schema] of Object.entries(spec.definitions)) {
      const data = schema.properties?.data;
      if (name.includes('WebResponse') && data !== undefined) {
        spec.definitions[name] = data as SwaggerSchema;
      }
    }
  }

  spec.basePath = '/';

  return spec;
}
