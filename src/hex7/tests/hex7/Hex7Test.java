package hex7;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.junit.jupiter.api.Assertions.fail;

import java.util.Arrays;
import java.util.HashSet;
import java.util.Random;
import java.util.Set;
import org.junit.jupiter.api.Test;

class Hex7Test {

  private static final Random rand = new Random();

  private static int bit(int value, int index) {
    return value >> index & 1;
  }

  private static int set(int value, int index, int bit) {
    return bit == 0
        ? value & ~(1 << index)
        : value | (1 << index);
  }

  @Test
  public void createAllHex7s() {
    for (int i = 0; i < 128; i++) {
      Hex7 hex = Hex7.of(i);
      assertEquals(0, hex.value() & (1 << 7), "Bit 7 is never set on a Hex7 value.");

      for (HexDir dir : HexDir.values()) {
        assertEquals(8, hex.couplingDirect1(dir).cardinality(),
            "8 neighbours at distance 1: " + dir);
        assertEquals(64, hex.couplingDirect2(dir).cardinality(),
            "64 neighbours at distance 2: " + dir);
        assertEquals(32, hex.couplingKnightsMove(dir).cardinality(),
            "32 neighbours at distance 2km: " + dir);
      }
    }
  }

  @Test()
  public void invalidValue() {
    try {
      Hex7.of(-1);
      fail("Expected IllegalArgumentException for negative values.");
    } catch (IllegalArgumentException expected) {
    }
  }

  @Test()
  public void invalidValue2() {
    try {
      Hex7.of(128);
      fail("Expected IllegalArgumentException for values > 127.");
    } catch (IllegalArgumentException expected) {
    }
  }

  @Test
  public void couples_R1() {
    for (int i = 0; i < 10_000; i++) {
      Hex7 hex1 = Hex7.of(rand.nextInt(128));
      int hex2value = rand.nextInt(128);

      // Reminder of Hex7 encoding schema:
      //   0 1
      //  5 6 2
      //   4 3
      // As long as bits 1, 6, 2, 3 of hex1 match bits 0, 5, 6, 4 of hex2,
      // hex2 should couple with hex1 when shifted 1 hexagon to the right.
      hex2value = set(hex2value, 0, bit(hex1.value(), 1));
      hex2value = set(hex2value, 5, bit(hex1.value(), 6));
      hex2value = set(hex2value, 6, bit(hex1.value(), 2));
      hex2value = set(hex2value, 4, bit(hex1.value(), 3));
      Hex7 hex2 = Hex7.of(hex2value);

      assertTrue(hex1.couplingDirect1(HexDir.R).get(hex2.value()),
          hex1 + "couples with" + hex2 + "in R direction.");
      assertTrue(hex2.couplingDirect1(HexDir.ZL).get(hex1.value()),
          hex2 + "couples with" + hex1 + "in L direction.");
    }
  }

  @Test
  public void couples_UR1() {
    for (int i = 0; i < 10_000; i++) {
      Hex7 hex1 = Hex7.of(rand.nextInt(128));
      int hex2value = rand.nextInt(128);

      //   0 1
      //  5 6 2
      //   4 3
      hex2value = set(hex2value, 5, bit(hex1.value(), 0));
      hex2value = set(hex2value, 4, bit(hex1.value(), 6));
      hex2value = set(hex2value, 6, bit(hex1.value(), 1));
      hex2value = set(hex2value, 3, bit(hex1.value(), 2));
      Hex7 hex2 = Hex7.of(hex2value);

      assertTrue(hex1.couplingDirect1(HexDir.UR).get(hex2.value()),
          hex1 + "couples with" + hex2 + "in UR direction.");
      assertTrue(hex2.couplingDirect1(HexDir.DL).get(hex1.value()),
          hex2 + "couples with" + hex1 + "in DL direction.");
    }
  }

  @Test
  public void couples_UL1() {
    for (int i = 0; i < 10_000; i++) {
      Hex7 hex1 = Hex7.of(rand.nextInt(128));
      int hex2value = rand.nextInt(128);

      //   0 1
      //  5 6 2
      //   4 3
      hex2value = set(hex2value, 4, bit(hex1.value(), 5));
      hex2value = set(hex2value, 6, bit(hex1.value(), 0));
      hex2value = set(hex2value, 3, bit(hex1.value(), 6));
      hex2value = set(hex2value, 2, bit(hex1.value(), 1));
      Hex7 hex2 = Hex7.of(hex2value);

      assertTrue(hex1.couplingDirect1(HexDir.UL).get(hex2.value()),
          hex1 + "couples with" + hex2 + "in UL direction.");
      assertTrue(hex2.couplingDirect1(HexDir.DR).get(hex1.value()),
          hex2 + "couples with" + hex1 + "in DR direction.");
    }
  }

  @Test
  public void couples_R2() {
    for (int i = 0; i < 10_000; i++) {
      Hex7 hex1 = Hex7.of(rand.nextInt(128));
      int hex2value = rand.nextInt(128);

      //   0 1
      //  5 6 2
      //   4 3
      hex2value = set(hex2value, 5, bit(hex1.value(), 2));
      Hex7 hex2 = Hex7.of(hex2value);

      assertTrue(hex1.couplingDirect2(HexDir.R).get(hex2.value()),
          hex1 + "couples with" + hex2 + "in 2 R direction.");
      assertTrue(hex2.couplingDirect2(HexDir.ZL).get(hex1.value()),
          hex2 + "couples with" + hex1 + "in 2 L direction.");
    }
  }

  @Test
  public void couples_UR2() {
    for (int i = 0; i < 10_000; i++) {
      Hex7 hex1 = Hex7.of(rand.nextInt(128));
      int hex2value = rand.nextInt(128);

      //   0 1
      //  5 6 2
      //   4 3
      hex2value = set(hex2value, 4, bit(hex1.value(), 1));
      Hex7 hex2 = Hex7.of(hex2value);

      assertTrue(hex1.couplingDirect2(HexDir.UR).get(hex2.value()),
          hex1 + "couples w/" + hex2 + "in 2 UR direction.");
      assertTrue(hex2.couplingDirect2(HexDir.DL).get(hex1.value()),
          hex2 + "couples w/" + hex1 + "in 2 DL direction.");
    }
  }

  @Test
  public void couples_UL2() {
    for (int i = 0; i < 10_000; i++) {
      Hex7 hex1 = Hex7.of(rand.nextInt(128));
      int hex2value = rand.nextInt(128);

      //   0 1
      //  5 6 2
      //   4 3
      hex2value = set(hex2value, 3, bit(hex1.value(), 0));
      Hex7 hex2 = Hex7.of(hex2value);

      assertTrue(hex1.couplingDirect2(HexDir.UL).get(hex2.value()),
          hex1 + "couples w/" + hex2 + "in 2 UL direction.");
      assertTrue(hex2.couplingDirect2(HexDir.DR).get(hex1.value()),
          hex2 + "couples w/" + hex1 + "in 2 DR direction.");
    }
  }

  @Test
  public void couples_R_DR() {
    for (int i = 0; i < 10_000; i++) {
      Hex7 hex1 = Hex7.of(rand.nextInt(128));
      int hex2value = rand.nextInt(128);

      // Knight's move R is actually R + DR:
      //   0 1
      //  5 6 2
      //   4 3
      // So hex1's 0, 5 should match hex2's 2, 3:
      hex2value = set(hex2value, 0, bit(hex1.value(), 2));
      hex2value = set(hex2value, 5, bit(hex1.value(), 3));
      Hex7 hex2 = Hex7.of(hex2value);

      assertTrue(hex1.couplingKnightsMove(HexDir.R).get(hex2.value()),
          hex1 + "couples with" + hex2 + "in Rkm direction.");
      assertTrue(hex2.couplingKnightsMove(HexDir.ZL).get(hex1.value()),
          hex2 + "couples with" + hex1 + "in Lkm direction.");
    }
  }

  @Test
  public void couples_UR_R() {
    for (int i = 0; i < 10_000; i++) {
      Hex7 hex1 = Hex7.of(rand.nextInt(128));
      int hex2value = rand.nextInt(128);

      //   0 1
      //  5 6 2
      //   4 3
      hex2value = set(hex2value, 5, bit(hex1.value(), 1));
      hex2value = set(hex2value, 4, bit(hex1.value(), 2));
      Hex7 hex2 = Hex7.of(hex2value);

      assertTrue(hex1.couplingKnightsMove(HexDir.UR).get(hex2.value()),
          hex1 + "couples w/" + hex2 + "in URkm direction.");
      assertTrue(hex2.couplingKnightsMove(HexDir.DL).get(hex1.value()),
          hex2 + "couples w/" + hex1 + "in DLkm direction.");
    }
  }

  @Test
  public void couples_UL_UR() {
    for (int i = 0; i < 10_000; i++) {
      Hex7 hex1 = Hex7.of(rand.nextInt(128));
      int hex2value = rand.nextInt(128);

      //   0 1
      //  5 6 2
      //   4 3
      hex2value = set(hex2value, 4, bit(hex1.value(), 0));
      hex2value = set(hex2value, 3, bit(hex1.value(), 1));
      Hex7 hex2 = Hex7.of(hex2value);

      assertTrue(hex1.couplingKnightsMove(HexDir.UL).get(hex2.value()),
          hex1 + "couples w/" + hex2 + "in ULkm direction.");
      assertTrue(hex2.couplingKnightsMove(HexDir.DR).get(hex1.value()),
          hex2 + "couples w/" + hex1 + "in DRkm direction.");
    }
  }

  @Test
  public void generateRotations() {
    Hex7 hex = Hex7.of(0x01);
    Set<Hex7> expected = new HashSet<>(Arrays.asList(
        hex, Hex7.of(0x02), Hex7.of(0x04), Hex7.of(0x08), Hex7.of(0x10), Hex7.of(0x20)));
    assertEquals(expected, hex.generateRotations(), "Rotations of:\n" + hex);
  }

  @Test
  public void generateRotations2() {
    Hex7 hex = Hex7.of(0x41);
    Set<Hex7> expected = new HashSet<>(Arrays.asList(
        hex, Hex7.of(0x42), Hex7.of(0x44), Hex7.of(0x48), Hex7.of(0x50), Hex7.of(0x60)));
    assertEquals(expected, hex.generateRotations(), "Rotations of:\n" + hex);
  }

  @Test
  public void generateRotations3() {
    Hex7 hex = Hex7.of(0x15);
    Set<Hex7> expected = new HashSet<>(Arrays.asList(hex, Hex7.of(0x2a)));
    assertEquals(expected, hex.generateRotations(), "Rotations of:\n" + hex);
  }

  @Test
  public void generateRotations4() {
    Hex7 hex = Hex7.of(0x49);
    Set<Hex7> expected = new HashSet<>(Arrays.asList(hex, Hex7.of(0x52), Hex7.of(0x64)));
    assertEquals(expected, hex.generateRotations(), "Rotations of:\n" + hex);
  }
}