package hex7;

import java.util.Arrays;
import java.util.BitSet;
import java.util.HashMap;
import java.util.Map;

/**
 * Toroidal surface with a hexagonal grid of size {@code M x N}. Stores a byte value in [0, 127] for
 * each hexagonal coordinate, representing the whole 7-hexagon neighborhood.
 */
public class Hex7Torus {

  private static final long[] BITS_128 = new long[]{-1L, -1L};
  /**
   * The hexagonal grid is stored as a flat array:
   * <pre>
   *   (0,0) (1,0) (2,0) (3,0) ...
   *      (0,1) (1,1) (2,1) (3,1) ...
   *   (0,2) (1,2) (2,2) (3,2) ...
   *      (0,3) (1,3) (2,3) (3,3) ...
   * </pre>
   * Each hexagon is like ⬢. The value of {@code 0xff} indicates that the hexagon is unset.
   */
  final byte[] array;
  private final int width, height;

  /**
   * Constructs a torus of size {@code width x height}. All hexagons are unset initially.
   */
  public Hex7Torus(int width, int height) {
    if (width < 0 || height < 0) {
      throw new IllegalArgumentException("Invalid height × width: " + height + " × " + width);
    }
    if (height % 2 != 0) {
      throw new IllegalArgumentException("Height must be divisible by 2, but was: " + height);
    }
    this.width = width;
    this.height = height;
    this.array = new byte[width * height];
    Arrays.fill(array, (byte) 0xff);
  }

  private Hex7Torus(int width, int height, byte[] array) {
    this.width = width;
    this.height = height;
    this.array = array;
  }

  Hex7Torus deepCopy() {
    return new Hex7Torus(
        width,
        height,
        Arrays.copyOf(array, array.length));
  }

  /**
   * Takes a {@link HexCoord} that is potentially out of bounds for this torus, wraps it around the
   * torus and returns a {@link HexCoord} that is always valid.
   */
  public HexCoord normalize(HexCoord coord) {
    return outOfBounds(coord)
        ? HexCoord.of((width + coord.x()) % width, (height + coord.y()) % height)
        : coord;
  }

  /**
   * @return {@code true} if the hexagon at {@code coord} is set.
   */
  public boolean isSet(HexCoord coord) {
    assert (!outOfBounds(coord));
    return array[index(coord)] != (byte) 0xff;
  }

  /**
   * @return {@link Hex7} at {@code coord} (which is expected to be "in bounds", no wrapping is
   * performed).
   */
  public Hex7 get(HexCoord coord) {
    assert (!outOfBounds(coord));
    if (!isSet(coord)) {
      throw new IllegalArgumentException("Hex7 at this HexCoord is unset: " + coord);
    }
    return get(index(coord));
  }

  /**
   * Convenience method for getting a {@link Hex7} by index.
   */
  public Hex7 get(int i) {
    return Hex7.of(array[i]);
  }

  /**
   * Sets the hexagon at {@code coord} (which is expected to be "in bounds", no wrapping is
   * performed) to the provided {@link Hex7}.
   */
  public void set(HexCoord coord, Hex7 hex) {
    assert (!outOfBounds(coord));
    set(index(coord), hex.value());
  }

  /**
   * Convenience method for setting a {@link Hex7} by index.
   */
  public void set(int i, byte value) {
    array[i] = value;
  }

  /**
   * Unsets the hexagon at {@code coord}, which is expected to be "in bounds" no wrapping is
   * performed.
   */
  public void unset(HexCoord coord) {
    assert (!outOfBounds(coord));
    int i = index(coord);
    array[i] = (byte) 0xff;
  }

  /**
   * Returns a {@link BitSet} containing all possible {@link Hex7}s that would couple (when put at
   * {@code coord}) with any potentially overlapping {@link Hex7}s already in this
   * {@link Hex7Torus}.
   */
  public BitSet available(HexCoord coord) {
    BitSet result = BitSet.valueOf(BITS_128); // All 128 bits set.

    // There are 18 Hex7s that need to be checked:
    // - 6 immediate neighbours.
    // - 6 neighbours that are 2 hexagons away on a straight line.
    // - 6 neighbours that are 2 hexagons away by a "knight's move".
    for (HexDir dir : HexDir.values()) {
      HexCoord c2 = normalize(coord.move(dir));
      if (isSet(c2)) {
        // Keep only the compatible Hex7s that can be put at 'coord', from another Hex7 POV.
        result.and(get(c2).couplingDirect1(dir.invert()));
      } else {
        // "Distance 2" neighbour is automatically satisfied if "distance 1" one was present.
        c2 = normalize(c2.move(dir)); // c2 now points to 2 hexagons away in the direction 'dir'.
        if (isSet(c2)) {
          result.and(get(c2).couplingDirect2(dir.invert()));
        }
      }
      // Handle knight's move.
      c2 = normalize(coord.move(dir).move(dir.next()));
      if (isSet(c2)) {
        result.and(get(c2).couplingKnightsMove(dir.invert()));
      }
    }
    return result;
  }

  /**
   * Calculates which {@link Hex7}s are present in the torus.
   */
  public BitSet calculateVisited() {
    BitSet visited = new BitSet(128);
    for (byte b : array) {
      if (b == (byte) 0xff) {
        continue;
      }
      visited.set(b);
    }
    return visited;
  }

  public boolean calculateSatisfiability(BitSet visited) {
    throw new RuntimeException("Not implemented.");
  }

  /**
   * Calculates the number of all unset hexagons in the torus.
   */
  public int calculateUnsetSize() {
    int result = 0;
    for (byte b : array) {
      if (b == (byte) 0xff) {
        result++;
      }
    }
    return result;
  }

  public boolean outOfBounds(HexCoord coord) {
    return coord.x() < 0 || coord.y() < 0 || coord.x() >= width || coord.y() >= height;
  }

  /**
   * Returns {@code true} if the hexagon at {@code coord} is unset and has at least 1 set
   * neighbour.
   */
  boolean isBoundary(HexCoord coord) {
    if (isSet(coord)) {
      return false;
    }
    for (HexDir dir : HexDir.values()) {
      if (isSet(normalize(coord.move(dir)))) {
        return true;
      }
    }
    return false;
  }

  /**
   * Converts a {@link HexCoord} to an index.
   */
  int index(HexCoord coord) {
    return coord.y() * width + coord.x();
  }

  /**
   * Converts an index to a {@link HexCoord}.
   */
  HexCoord coord(int index) {
    return HexCoord.of(index % width, index / width);
  }

  /**
   * Returns a mapping from {@link Hex7} to a count of how many times it appeared in the torus.
   * <p>
   * Note: contrary to a properly populated {@link Hex7Torus}, this method assumes every hexagon
   * contains either 0, 1 or {@code 0xff} for unpopulated hexagons.
   */
  public Map<Hex7, Integer> verify() {
    Map<Hex7, Integer> counts = new HashMap<>();
    for (int i = 0; i < array.length; i++) {
      if (array[i] == (byte) 0xff) {
        // There is no full Hex7 here.
        continue;
      }
      if (array[i] != 0 && array[i] != 1) {
        throw new IllegalStateException("Verifier expects only 0, 1 or 0xff as hexagon values.");
      }
      // Center is encoded as bit 6.
      int value = array[i] == 0 ? 0 : (1 << 6);

      HexCoord center = coord(i);
      for (HexDir dir : HexDir.values()) {
        int value2 = array[index(normalize(center.move(dir)))];
        if (value2 == (byte) 0xff) {
          // There no fully populated Hex7 here.
          value = -1;
          break;
        }
        value |= value2 == 0 ? 0 : (1 << dir.ordinal());
      }

      if (value == -1) {
        continue;
      }

      Hex7 hex = Hex7.of(value & 0xff);
      counts.put(hex, counts.getOrDefault(hex, 0) + 1);
    }
    return counts;
  }

  @Override
  public String toString() {
    // xaa xbb xcc
    //   xdd xee xff ...
    StringBuilder result = new StringBuilder();
    // 1 1 1
    //  1 1 1 ...
    StringBuilder result2 = new StringBuilder();
    for (int i = 0; i < array.length; i++) {
      if (i >= width && i % width == 0) {
        result.append("\n");
        result2.append("\n");
        if ((i / width) % 2 == 1) {
          result.append("  ");
          result2.append(" ");
        }
      }
      result.append(array[i] == (byte) 0xff ? "... " : String.format("x%02x ", array[i]));

      String value = ". ";
      if (array[i] != (byte) 0xff) {
        value = (array[i] >> 6 & 1) + " ";
      } else if (isBoundary(coord(i))) {
        // Find the populated neighbour.
        HexDir dir = HexDir.UL;
        while (!isSet(normalize(coord(i).move(dir)))) {
          dir = dir.next();
        }
        value = get(normalize(coord(i).move(dir))).bit(dir.invert().ordinal()) == 0 ? "O " : "I ";
      }
      result2.append(value);
    }
    return result + "\n\n" + result2;
  }

  public int width() {
    return width;
  }

  public int height() {
    return height;
  }
}
