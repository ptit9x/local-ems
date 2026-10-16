import { IsOptional, IsDateString, IsIn, IsInt, Max, Min } from 'class-validator';
import { Type } from 'class-transformer';

export class TelemetryQueryDto {
  @IsOptional() @IsDateString() from?: string;

  @IsOptional() @IsDateString() to?: string;

  @IsOptional()
  @Type(() => Number) @IsInt() @Min(1) @Max(5000)
  limit?: number;

  @IsOptional() @IsIn(['raw', '1m', '5m', '1h', '1d']) bucket?: string;
}
