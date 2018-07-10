package n3

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

// Sample hardcoded queries; for use with NSW Digtial Classroom project

// return the GUIDs of all students studying the Learning Area at the Year Level.
// involves joining StudentPersonal, TeachingGroup, TimeTableSubject, SchoolCourseInfo.
// Presupposes that Learning Area is encoded as SchoolCourseInfo/SubjectAreaList/SubjectArea/Code
// Does not attempt any filtering by school: Assumes single school in the hexastore.
func (hx *Hexastore) KLAtoStudentQuery(kla string, yrlvl string) ([]string, error) {
	ret := make([]string, 0)
	teachinggroupIds, err := hx.KLAtoTeachingGroupQuery(kla, yrlvl)
	if err != nil {
		return ret, err
	}
	log.Printf("hx.KLAtoStudentQuery: teachinggroupIds: %#v\n", teachinggroupIds)
	for _, x := range teachinggroupIds {
		students, err := hx.GetTuples(fmt.Sprintf("c:\"SIF\" s:\"%s\" p:\"TeachingGroup.StudentList.TeachingGroupStudent", x))
		if err != nil {
			log.Println(err)
			return ret, err
		}
		log.Printf("hx.KLAtoStudentQuery: students: %#v\n", students)
		for _, y := range students {
			if strings.HasSuffix(y.Predicate, ".StudentPersonalRefId") {
				ret = append(ret, y.Object)
			}
		}
	}
	log.Printf("hx.KLAtoStudentQuery: ret: %#v\n", ret)
	return ret, nil
}

func (hx *Hexastore) KLAtoTeacherQuery(kla string, yrlvl string) ([]string, error) {
	ret := make([]string, 0)
	teachinggroupIds, err := hx.KLAtoTeachingGroupQuery(kla, yrlvl)
	if err != nil {
		return ret, err
	}
	for _, x := range teachinggroupIds {
		students, err := hx.GetTuples(fmt.Sprintf("c:\"SIF\" s:\"%s\" p:\"TeachingGroup.TeacherList.TeachingGroupTeacher", x))
		if err != nil {
			return ret, err
		}
		for _, y := range students {
			if strings.HasSuffix(y.Predicate, ".StaffPersonalRefId") {
				ret = append(ret, y.Object)
			}
		}
	}
	return ret, nil
}

func (hx *Hexastore) KLAtoTeachingGroupQuery(kla string, yrlvl string) ([]string, error) {
	schoolcourseinfoIds := make([]string, 0)
	timetablesubjectIds := make([]string, 0)
	ret := make([]string, 0)
	curr, err := strconv.Atoi(yrlvl)
	if err != nil {
		log.Println(err)
		return ret, err
	}

	schoolcourseinfos, err := hx.GetTuples(`c:"SIF" p:"SchoolCourseInfo.SubjectAreaList`)
	if err != nil {
		log.Println(err)
		return ret, err
	}
	for _, x := range schoolcourseinfos {
		if strings.HasSuffix(x.Predicate, ".Code") {
			if x.Object == kla {
				schoolcourseinfoIds = append(schoolcourseinfoIds, x.Subject)
			}
		}
	}
	//log.Printf("KLAtoTeachingGroupQuery: schoolcourseinfoIds, %#v\n", schoolcourseinfoIds)
	for _, x := range schoolcourseinfoIds {
		timetablesubjects, err := hx.GetTuples(fmt.Sprintf("c:\"SIF\" p:\"TimeTableSubject.SchoolCourseInfoRefId\" o:\"%s\" ", x))
		if err != nil {
			log.Println(err)
			return ret, err
		}
		//log.Printf("KLAtoTeachingGroupQuery: timetablesubjects, %#v\n", timetablesubjects)
		for _, y := range timetablesubjects {
			found := false
			yrlvls, err := hx.GetTuples(fmt.Sprintf("c:\"SIF\" s:\"%s\" p:\"TimeTableSubject.AcademicYear.Code\" ", y.Subject))
			if err != nil {
				log.Println(err)
				return ret, err
			}
			for _, z := range yrlvls {
				found = found || z.Object == yrlvl
			}
			if !found {
				yrlvls1, err := hx.GetTuples(fmt.Sprintf("c:\"SIF\" s:\"%s\" p:\"TimeTableSubject.AcademicYearRange.Start.Code\" ", y.Subject))
				if err != nil {
					log.Println(err)
					return ret, err
				}
				yrlvls2, err := hx.GetTuples(fmt.Sprintf("c:\"SIF\" s:\"%s\" p:\"TimeTableSubject.AcademicYearRange.End.Code\" ", y.Subject))
				if err != nil {
					log.Println(err)
					return ret, err
				}
				if len(yrlvls1) == 1 && len(yrlvls2) == 1 {
					start, err := strconv.Atoi(yrlvls1[0].Object)
					if err != nil {
						log.Println(err)
						return ret, err
					}
					end, err := strconv.Atoi(yrlvls2[0].Object)
					if err != nil {
						return ret, err
					}
					found = start <= curr && end >= curr
				}
			}
			if found {
				timetablesubjectIds = append(timetablesubjectIds, y.Subject)
			}
		}
	}
	//log.Printf("KLAtoTeachingGroupQuery: timetablesubjectIds, %#v\n", timetablesubjectIds)
	for _, x := range timetablesubjectIds {
		//teachinggroups, err := hx.GetTuples(fmt.Sprintf("c:\"SIF\" o:\"%s\" p:\"TeachingGroup.TimeTableSubjectRefId\" ", x))
		teachinggroups, err := hx.GetTuples(fmt.Sprintf("c:\"SIF\" p:\"TeachingGroup.TimeTableSubjectRefId\" o:\"%s\" ", x))
		if err != nil {
			log.Println(err)
			return ret, err
		}
		for _, y := range teachinggroups {
			ret = append(ret, y.Subject)
		}
	}
	return ret, nil
}
