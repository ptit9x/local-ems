import { z } from 'zod';

export const TelemetryQuerySchema = z.object({
  from: z.string().datetime().optional(),
  to: z.string().datetime().optional(),
  limit: z.coerce.number().int().min(1).max(5000).optional(),
  bucket: z.enum(['raw', '1m', '5m', '1h', '1d']).optional(),
});

export type TelemetryQueryDto = z.infer<typeof TelemetryQuerySchema>;
