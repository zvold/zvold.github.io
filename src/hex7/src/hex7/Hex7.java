package hex7;

import java.util.BitSet;
import java.util.HashSet;
import java.util.Objects;
import java.util.Set;

/**
 * Represents a 7-hex neighbourhood on a hexagonal grid, where each hexagon ⬢ can be 0 or 1.
 * <pre>
 *  0 1
 * 1 0 0
 *  0 1
 * </pre>
 */
public class Hex7 {

  /**
   * Lazily populated cache of all 128 possible {@link Hex7}s.
   */
  private static final Hex7[] cache = new Hex7[128];

  /**
   * The 7 adjacent hexagons are encoded as 7 bits stored in a byte:
   * <pre>
   *   0 1
   *  5 6 2
   *   4 3
   * </pre>
   * Note that the possible values are in [0, 127], and the highest bit is never set.
   */
  private final byte value;

  /**
   * For a given {@link Hex7}, stores all possible non-conflicting neighbouring {@link Hex7}s at
   * manhattan distance of 1, as 6 bitmasks, one for each {@link HexDir}.
   * <p>
   * For example, this {@link Hex7} allows 8 {@link Hex7}s in the {@link HexDir#R} direction:
   * <pre>
   *    a B         B x
   *   c D E  -->  D E y
   *    f G         G z
   * </pre>,
   * where x, y, z are arbitrary bits, and B, D, E, G hexagons on the left belong to the original
   * {@link Hex7}. The {@code neighbours[HexDir.R]} bitmask will have 8 (and only 8) corresponding
   * bits set.
   */
  private final BitSet[] couplingDirect1;

  /**
   * For a given {@link Hex7}, stores all possible non-conflicting neighbouring {@link Hex7}s at
   * manhattan distance of 2 in straight line, as 6 bitmasks, one for each {@link HexDir}.
   * <p>
   * For example, this {@link Hex7} allows 2^6 {@link Hex7}s in the {@link HexDir#R} direction:
   * <pre>
   *    a b         x y
   *   c d E  -->  E z s
   *    f g         t p
   * </pre>,
   * where x, y, z, s, t, p are arbitrary bits, E hexagon on the left belongs to the original
   * {@link Hex7}. The {@code neighbours[HexDir.R]} bitmask will have 64 (and only 64) corresponding
   * bits set.
   */
  private final BitSet[] couplingDirect2;

  /**
   * For a given {@link Hex7}, stores all possible non-conflicting neighbouring {@link Hex7}s at
   * manhattan distance of 2 in a "knight's move" (two moves made in sequential {@link HexDir}s).
   * The configurations are stored as 6 bitmasks, one for each {@link HexDir}.
   * <p>
   * For example, this {@link Hex7} allows 2^5 {@link Hex7}s in the {@link HexDir#UR} +
   * {@link HexDir#R} direction:
   * <pre>
   *                x y
   *    a B        B z s
   *   c d E  -->   E t
   *    f g
   * </pre>,
   * where x, y, z, s, t are arbitrary bits, and B, E hexagons on the left belongs to the original
   * {@link Hex7}. The {@code neighbours[HexDir.R]} bitmask will have 32 (and only 32) corresponding
   * bits set.
   */
  private final BitSet[] couplingKnightsMove;

  private Hex7(byte value) {
    this.value = value;

    couplingDirect1 = new BitSet[6];
    couplingDirect2 = new BitSet[6];
    couplingKnightsMove = new BitSet[6];
    for (int i = 0; i < 6; i++) {
      couplingDirect1[i] = new BitSet(128);
      couplingDirect2[i] = new BitSet(128);
      couplingKnightsMove[i] = new BitSet(128);
    }

    for (HexDir dir : HexDir.values()) {
      int i = dir.ordinal();
      for (int hex = 0; hex < 128; hex++) {
        // Pre-populate bitmasks for neighbours at manhattan distance 1.
        if (bit(hex, (i + 4) % 6) == bit((i + 5) % 6)
            && bit(hex, (i + 3) % 6) == bit(6)
            && bit(hex, 6) == bit(i)
            && bit(hex, (i + 2) % 6) == bit((i + 1) % 6)) {
          couplingDirect1[i].set(hex);
        }
        // Pre-populate bitmasks for neighbours at manhattan distance of 2 on a straight line.
        if (bit(hex, (3 + i) % 6) == bit(i)) {
          couplingDirect2[i].set(hex);
        }
        // Pre-populate bitmasks for neighbours at manhattan distance of 2 on a "knight's move".
        if (bit(hex, (4 + i) % 6) == bit(i)
            && bit(hex, (3 + i) % 6) == bit((1 + i) % 6)) {
          couplingKnightsMove[i].set(hex);
        }
      }
    }
  }

  /**
   * Constructs a {@link Hex7} corresponding to the given 7-hexagon neighbourhood, encoded as a
   * single {@code value} using this bit order:
   * <pre>
   *   0 1
   *  5 6 2
   *   4 3
   * </pre>
   */
  public static Hex7 of(int value) {
    if (value < 0 || value > 127) {
      throw new IllegalArgumentException("Invalid Hex7 value: " + value);
    }
    if (cache[value] == null) {
      cache[value] = new Hex7((byte) value);
    }
    return cache[value];
  }

  /**
   * Returns 0 or 1 depending on whether bit at {@code index} is set in {@code value}.
   */
  private static int bit(int value, int index) {
    return (value >> index & 1) != 0 ? 1 : 0;
  }

  /**
   * Returns the bitmask encoding possible neighbours at manhattan distance of 1, in the direction
   * {@code dir}.<p> Callers must not modify the returned {@link BitSet}.
   */
  BitSet couplingDirect1(HexDir dir) {
    return couplingDirect1[dir.ordinal()];
  }

  /**
   * Returns the bitmask encoding possible neighbours at manhattan distance of 2, in the direction
   * {@code dir}.<p> Callers must not modify the returned {@link BitSet}.
   */
  BitSet couplingDirect2(HexDir dir) {
    return couplingDirect2[dir.ordinal()];
  }

  /**
   * Returns the bitmask encoding possible neighbours at manhattan distance of 2, in the direction
   * {@code dir} on "knight's move".<p> Callers must not modify the returned {@link BitSet}.
   */
  BitSet couplingKnightsMove(HexDir dir) {
    return couplingKnightsMove[dir.ordinal()];
  }

  public byte value() {
    return value;
  }

  /**
   * Returns 0 or 1 depending on whether bit at {@code index} is set in {@link Hex7#value}.
   */
  int bit(int index) {
    return bit(value, index);
  }

  @Override
  public String toString() {
    return String.format("\n  %d %d \n %d %d %d\n  %d %d\n",
        bit(0), bit(1), bit(5), bit(6), bit(2), bit(4), bit(3));
  }

  @Override
  public boolean equals(Object o) {
    if (this == o) {
      return true;
    }
    if (!(o instanceof Hex7 hex7)) {
      return false;
    }
    return value == hex7.value;
  }

  @Override
  public int hashCode() {
    return Objects.hashCode(value);
  }

  /**
   * Returns unique rotations of this {@link Hex7}.
   */
  public Set<Hex7> generateRotations() {
    Set<Hex7> result = new HashSet<>();
    result.add(this);
    int v = value;
    for (int i = 1; i < 6; i++) {
      v <<= 1;
      v &= 0b0111_1110;
      v |= (v >> 6 & 1);
      v &= 0b0011_1111;
      v |= (value & 0b0100_0000);
      result.add(Hex7.of(v));
    }
    return result;
  }
}
