package hex7;

import java.util.function.Function;

/**
 * Represents 6 directions on the hexagonal grid, where each hexagon is like ⬢.
 */
public enum HexDir {
  // Note: order is important, and matches the encoding of Hex7 values (except 6 for the center):
  //   0 1
  //  5 6 2
  //   4 3
  UL, UR, R, DR, DL, ZL;

  public static void forEach(Function<HexDir, Boolean> func) {
    for (HexDir dir : HexDir.values()) {
      if (!func.apply(dir)) {
        break;
      }
    }
  }

  public HexDir invert() {
    return HexDir.values()[(ordinal() + 3) % 6];
  }

  public HexDir next() {
    return HexDir.values()[(ordinal() + 1) % 6];
  }
}
