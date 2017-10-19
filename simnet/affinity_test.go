package simnet

import "testing"

func TestIsolate(t *testing.T) {
	t.Run("Within same city", func(t *testing.T) {
		isolation := Isolate(NewPointFromCity("北京市"), NewPointFromCity("北京市"))
		if isolation > 0 {
			t.Errorf("expected 0, got %v", isolation)
		}
	})

	t.Run("Within same province less than not", func(t *testing.T) {
		for _, a := range names {
			aCity := cities[a]
			for _, b := range names {
				bCity := cities[b]
				for _, c := range names {
					cCity := cities[c]
					if aCity.Province == bCity.Province && bCity.Province != cCity.Province {
						c1, c2, c3 := NewPointFromCity(aCity.Name), NewPointFromCity(bCity.Name), NewPointFromCity(cCity.Name)
						v1 := Isolate(c1, c2)
						v2 := Isolate(c2, c3)
						if v1 >= v2 {
							t.Errorf("expected %v-%v less than %v-%v, got opposite", c1.City, c2.City, c2.City, c3.City)
						}
					}
				}
			}
		}
	})

	t.Run("Within same district less than not", func(t *testing.T) {
		for _, a := range names {
			aCity := cities[a]
			for _, b := range names {
				bCity := cities[b]
				for _, c := range names {
					cCity := cities[c]
					if aCity.District == bCity.District && bCity.District != cCity.District {
						c1, c2, c3 := NewPointFromCity(aCity.Name), NewPointFromCity(bCity.Name), NewPointFromCity(cCity.Name)
						v1 := Isolate(c1, c2)
						v2 := Isolate(c2, c3)
						if v1 >= v2 {
							t.Errorf("expected %v-%v less than %v-%v, got opposite", c1.City, c2.City, c2.City, c3.City)
						}
					}
				}
			}
		}
	})
}

func TestAffinity(t *testing.T) {
	var points []Point
	for i := 0; i < len(names); i++ {
		points = append(points, NewPoint(i))
	}
	for _, z := range NewAffinity(points) {
		if z.A.City.Province == z.B.City.Province && z.A.ISP == z.B.ISP {
			// All IDCs within same province and same ISP construct a connected graph.
			if z.PacketLoss == 100 {
				t.Errorf("expected reachable from %v to %v, got with 100%% packet loss", z.A.City, z.B.City)
			}
		}
		if z.A.City.Province != z.B.City.Province && z.A.ISP == z.B.ISP {
			// Existing an IDC on capital of down region can reach to an IDC on capital of up region with same ISP.
			if z.A.City.District == z.B.City.District && secondClassInfluxes[z.B.City.Name] ||
				firstClassCores[z.B.City.Name] {
				if z.PacketLoss == 100 {
					t.Errorf("expected reachable from %v to %v, got with 100%% packet loss", z.A.City, z.B.City)
				}
			}
		}

		if z.A.City.Province == z.B.City.Province && z.A.ISP != z.B.ISP {
			// Existing an IDC can reach to one of the base ISPs for any ISP within same province. Base is the minimal set of ISPs that our management machine built on.
			if baseISPs[int(z.B.ISP)] && z.PacketLoss == 100 {
				t.Errorf("expected reachable from %v to %v, got with 100%% packet loss", z.A.City, z.B.City)
			}
		}
	}
}
