package hex7;

import static org.junit.jupiter.api.Assertions.assertEquals;

import java.util.EnumSet;
import org.junit.jupiter.api.Test;

class HexDirTest {

  @Test
  public void testInvert() {
    assertEquals(HexDir.R, HexDir.ZL.invert(), "R <-> L");
    assertEquals(HexDir.ZL, HexDir.R.invert(), "L <-> R");

    assertEquals(HexDir.UR, HexDir.DL.invert(), "UR <-> DL");
    assertEquals(HexDir.DL, HexDir.UR.invert(), "DL <-> UR");

    assertEquals(HexDir.DR, HexDir.UL.invert(), "DR <-> UL");
    assertEquals(HexDir.UL, HexDir.DR.invert(), "UL <-> DR");
  }

  @Test
  public void testNext() {
    assertEquals(HexDir.UR, HexDir.UL.next(), "UL -> UR");
    assertEquals(HexDir.UL, HexDir.ZL.next(), "L -> UL");
  }

  @Test
  public void randomOrder() {
    EnumSet<HexDir> visited = EnumSet.noneOf(HexDir.class);
    HexDir dir = HexDir.R;
    for (int d = 0; d < 6; d++, dir = dir.next()) {
      visited.add(dir);
    }
    assertEquals(6, visited.size(), "All directions visited.");
  }
}