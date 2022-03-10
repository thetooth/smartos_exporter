package collector

import (
	"testing"
)

func TestSMART(t *testing.T) {
	c, err := NewGZDiskSMARTExporter()
	if err != nil {
		t.Fatal(err)
	}
	input := `smartctl 7.1 2019-12-30 r5022 [x86_64-sun-solaris2.11] (local build)
Copyright (C) 2002-19, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF READ SMART DATA SECTION ===
SMART Attributes Data Structure revision number: 10
Vendor Specific SMART Attributes with Thresholds:
ID# ATTRIBUTE_NAME          FLAG     VALUE WORST THRESH TYPE      UPDATED  WHEN_FAILED RAW_VALUE
	1 Raw_Read_Error_Rate     0x000f   083   067   044    Pre-fail  Always       -       184202112
	3 Spin_Up_Time            0x0003   092   092   000    Pre-fail  Always       -       0
	4 Start_Stop_Count        0x0032   100   100   020    Old_age   Always       -       6
	5 Reallocated_Sector_Ct   0x0033   100   100   010    Pre-fail  Always       -       0
	7 Seek_Error_Rate         0x000f   066   060   045    Pre-fail  Always       -       4375124
	9 Power_On_Hours          0x0032   100   100   000    Old_age   Always       -       216
	10 Spin_Retry_Count        0x0013   100   100   097    Pre-fail  Always       -       0
	12 Power_Cycle_Count       0x0032   100   100   020    Old_age   Always       -       6
	18 Unknown_Attribute       0x000b   100   100   050    Pre-fail  Always       -       0
187 Reported_Uncorrect      0x0032   100   100   000    Old_age   Always       -       0
188 Command_Timeout         0x0032   100   100   000    Old_age   Always       -       0
190 Airflow_Temperature_Cel 0x0022   067   048   040    Old_age   Always       -       33 (Min/Max 29/44)
192 Power-Off_Retract_Count 0x0032   100   100   000    Old_age   Always       -       4
193 Load_Cycle_Count        0x0032   100   100   000    Old_age   Always       -       106
194 Temperature_Celsius     0x0022   033   052   000    Old_age   Always       -       33 (0 20 0 0 0)
195 Hardware_ECC_Recovered  0x001a   083   067   000    Old_age   Always       -       184202112
197 Current_Pending_Sector  0x0012   100   100   000    Old_age   Always       -       0
198 Offline_Uncorrectable   0x0010   100   100   000    Old_age   Offline      -       0
199 UDMA_CRC_Error_Count    0x003e   200   200   000    Old_age   Always       -       0
200 Multi_Zone_Error_Rate   0x0023   100   100   001    Pre-fail  Always       -       0
240 Head_Flying_Hours       0x0000   100   253   000    Old_age   Offline      -       33 (175 70 0)
241 Total_LBAs_Written      0x0000   100   253   000    Old_age   Offline      -       4287377051
242 Total_LBAs_Read         0x0000   100   253   000    Old_age   Offline      -       4201974967
`

	err = c.parseSmartCtlOutput("", "", input)
	if err != nil {
		t.Fatal(err)
	}
}
