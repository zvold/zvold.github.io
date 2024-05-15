package hex7;

import java.util.Objects;

/**
 * Coordinate on a hexagonal grid, where each hexagon is like ⬢:
 * <pre>
 *   (0,0) (1,0) (2,0) (3,0) ...
 *      (0,1) (1,1) (2,1) (3,1) ...
 *   (0,2) (1,2) (2,2) (3,2) ...
 *      (0,3) (1,3) (2,3) (3,3) ...
 * </pre>
 */
public class HexCoord {

  private final int x, y;

  private HexCoord(int x, int y) {
    this.x = x;
    this.y = y;
  }

  public static HexCoord of(int x, int y) {
    return new HexCoord(x, y);
  }

  public int x() {
    return x;
  }

  public int y() {
    return y;
  }

  public HexCoord move(HexDir dir) {
    return switch (dir) {
      case UL -> HexCoord.of(x - (y % 2 == 0 ? 1 : 0), y - 1);
      case UR -> HexCoord.of(x + (y % 2 == 0 ? 0 : 1), y - 1);
      case R -> HexCoord.of(x + 1, y);
      case DR -> HexCoord.of(x + (y % 2 == 0 ? 0 : 1), y + 1);
      case DL -> HexCoord.of(x - (y % 2 == 0 ? 1 : 0), y + 1);
      case ZL -> HexCoord.of(x - 1, y);
    };
  }

  @Override
  public boolean equals(Object o) {
    if (this == o) {
      return true;
    }
    if (!(o instanceof HexCoord hexCoord)) {
      return false;
    }
    return x == hexCoord.x && y == hexCoord.y;
  }

  @Override
  public int hashCode() {
    return Objects.hash(x, y);
  }

  @Override
  public String toString() {
    return "<" + x + "," + y + ">";
  }
}
