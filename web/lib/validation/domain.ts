import { z } from "zod";

export const createDomainSchema = z.object({
  hostname: z
    .string()
    .min(1, "Hostname is required")
    .regex(
      /^([a-z0-9]([a-z0-9-]*[a-z0-9])?\.)+[a-z]{2,}$/,
      "Must be a valid hostname (e.g. app.example.com)"
    ),
  baseDomain: z.string().min(1, "Base domain is required"),
  provider: z.enum(["manual", "duckdns"]),
  routingType: z.enum(["proxy", "redirect"]),
  tlsEnabled: z.boolean(),
  projectServiceId: z.string().optional(),
});

export type CreateDomainValues = z.infer<typeof createDomainSchema>;
