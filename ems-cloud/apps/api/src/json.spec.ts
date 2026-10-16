import { enableBigIntJson } from './json';

describe('enableBigIntJson', () => {
  it('serializes Prisma bigint identifiers as lossless decimal strings', () => {
    enableBigIntJson();
    expect(JSON.stringify({ id: BigInt('9007199254740993') })).toBe('{"id":"9007199254740993"}');
  });
});
