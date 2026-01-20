import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import { isTokenExpired } from './utils/auth'
import SearchBar from './search-bar.tsx'
import DiseaseTable from './disease.tsx'
import DrugTable from './drug.tsx'

interface PatientData {
  username: string,
  email: string,
  national_id: string,
}

interface Disease {
  id: string,
  rx_norm_id: string,
  name: string,
}

function Prescribe() {
  const [patientData, setPatientData] = useState<PatientData>({
    username: "",
    email: "",
    national_id: "",
  });

  const [diseaseData, setDiseaseData] = useState<Disease[]>([]);
  const [diseasePage, setDiseasePage] = useState<int>(1);
  const [diseaseTotalPages, setDiseaseTotalPages] = useState<int>(1);

  const [recDrugData, setRecDrugData] = useState<Drug[]>([]);
  const [recDrugPage, setRecDrugPage] = useState<int>(1);
  const [recDrugTotalPages, setRecDrugTotalPages] = useState<int>(1);

  const [manualDrugData, setManualDrugData] = useState<Drug[]>([]);
  const [manualDrugPage, setManualDrugPage] = useState<int>(1);
  const [manualDrugTotalPages, setManualDrugTotalPages] = useState<int>(1);

  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [diseaseErrorMessage, setDiseaseErrorMessage] = useState<string | null>(null);
  const [recDrugErrorMessage, setRecDrugErrorMessage] = useState<string | null>(null);
  const [manualDrugErrorMessage, setManualDrugErrorMessage] = useState<string | null>(null);

  let [patientSearchString, setPatientSearchString] = useState<string>("");
  let [diseaseSearchString, setDiseaseSearchString] = useState<string>("");
  let [drugSearchString, setDrugSearchString] = useState<string>("");

  let [selectedDisease, setSelectedDisease] = useState<Disease | null>(null);
  let [selectedDrugs, setSelectedDrugs] = useState<Drug[]>([]);

  const [recOrderBy, setRecOrderBy] = useState<"name" | "substance">("name");
  const [recOrderMethod, setRecOrderMethod] = useState<"asc" | "desc">("asc");

  const [manualOrderBy, setManualOrderBy] = useState<"name" | "substance">("name");
  const [manualOrderMethod, setManualOrderMethod] = useState<"asc" | "desc">("asc");

  let navigate = useNavigate();

  const handleDiseaseSelect = (disease: Disease) => {
    setSelectedDisease(disease); 
    recommendDrugs(disease, 1, recOrderBy, recOrderMethod);
  }

  const handleRecommendedDrugSelect = (drugs: Drug[]) => {
    setSelectedDrugs(drugs);
  };

  const toggleRecSort = (field: "name" | "substance") => {
    const newOrderBy = recOrderBy === field ? recOrderBy : field;
    const newOrderMethod =
      recOrderBy === field
        ? recOrderMethod === "asc"
        ? "desc"
        : "asc"
      : "asc";

    setRecOrderBy(newOrderBy);
    setRecOrderMethod(newOrderMethod);

    if (selectedDisease) {
      recommendDrugs(selectedDisease, 1, newOrderBy, newOrderMethod);
    }
  };

  const toggleManualSort = (field: "name" | "substance") => {
    const newOrderBy = manualOrderBy === field ? manualOrderBy : field;
    const newOrderMethod =
      manualOrderBy === field
        ? manualOrderMethod === "asc"
        ? "desc"
        : "asc"
      : "asc";

    setManualOrderBy(newOrderBy);
    setManualOrderMethod(newOrderMethod);

    onDrugSearch(manualDrugPage, newOrderBy, newOrderMethod);
  };

  const onPatientSearch = () => {
    let jwt = localStorage.getItem("jwt"); 
    if (isTokenExpired(jwt)) {
      navigate("/");
    }

    if (jwt === null) {
      throw new Error("Authentication failed");
    } else {
      fetch(`http://localhost:4000/med/patient/${patientSearchString}`, {
        method: "GET",
        headers: {
          "accept": "application/json",
          "Authorization": jwt,
        }
      }) 
      .then((res) => {
        if (res.status === 400) {
          throw new Error("Invalid National ID"); 
        } else if (res.status === 401 || res.status === 403) {
          throw new Error("You do not have permission to perform this action"); 
        } else if (res.status === 500) {
          throw new Error("Something went wrong"); 
        }
        return res.json();
      })
      .then((data) => {
        setPatientData(data); 
        setErrorMessage(null);
      })
      .catch((err) => {
        setErrorMessage(err.message);
      })
    }
  };

  const onDiseaseSearch = (page = 1) => {
    let jwt = localStorage.getItem("jwt"); 
    if (isTokenExpired(jwt)) {
      navigate("/");
    }

    if (jwt === null) {
      throw new Error("Authentication failed");
    } else {
      fetch(`http://localhost:4000/med/disease/${diseaseSearchString}?page=${page}`, {
        method: "GET",
        headers: {
          "accept": "application/json",
          "Authorization": jwt,
        }
      }) 
      .then((res) => {
        if (res.status === 400) {
          throw new Error("Invalid parameters"); 
        } else if (res.status === 401 || res.status === 403) {
          throw new Error("You do not have permission to perform this action"); 
        } else if (res.status === 404) {
          throw new Error("Disease not found"); 
        } else if (res.status === 500) {
          throw new Error("Something went wrong"); 
        }
        return res.json();
      })
      .then((data) => {
        setDiseaseData(data.data);
        setDiseasePage(data.page);
        setDiseaseTotalPages(data.total_pages);
        setDiseaseErrorMessage(null);
      })
      .catch((err) => {
        setDiseaseErrorMessage(err.message);
        setDiseasePage(1);
      })
    }
  };

  const recommendDrugs = (
    disease: Disease,
    page: number = 1,
    orderBy: "name" | "substance",
    orderMethod: "asc" | "desc"
  ) => {
    let jwt = localStorage.getItem("jwt"); 
    if (isTokenExpired(jwt)) {
      navigate("/");
    }
    if(jwt === null) {
      throw new Error("Authentication failed");
    } else {
      fetch(`http://localhost:4000/med/drug/recommend/${disease.rx_norm_id}?page=${page}&orderBy=${orderBy}&orderMethod=${orderMethod}`, {
        method: "GET",
        headers: {
          "accept": "application/json",
          "Authorization": jwt,
        }
      })
      .then((res) => {
        if (res.status === 400) {
          throw new Error("Invalid parameters"); 
        } else if (res.status === 401 || res.status === 403) {
          throw new Error("You do not have permission to perform this action"); 
        } else if (res.status === 404) {
          throw new Error("Drug not found"); 
        } else if (res.status === 500) {
          throw new Error("Something went wrong"); 
        }
        return res.json();
      })
      .then((data) => {
        setRecDrugData(data.data);
        setRecDrugPage(data.page);
        setRecDrugTotalPages(data.total_pages);
        setRecDrugErrorMessage(null);
      })
      .catch((err) => { 
        setRecDrugData([]);
        setRecDrugPage(1);
        setRecDrugTotalPages(1);
        setRecDrugErrorMessage(err.message);
      })
    };
  }

  const onDrugSearch = (
    page: number = 1,
    orderBy: "name" | "substance",
    orderMethod: "asc" | "desc"
  ) => {
    let jwt = localStorage.getItem("jwt"); 
    if (isTokenExpired(jwt)) {
      navigate("/");
    }
    if(jwt === null) {
      throw new Error("Authentication failed");
    } else {
      fetch(`http://localhost:4000/med/drug/${drugSearchString}?page=${page}&orderBy=${orderBy}&orderMethod=${orderMethod}`, {
        method: "GET",
        headers: {
          "accept": "application/json",
          "Authorization": jwt,
        }
      })
      .then((res) => {
        if (res.status === 400) {
          throw new Error("Invalid parameters"); 
        } else if (res.status === 401 || res.status === 403) {
          throw new Error("You do not have permission to perform this action"); 
        } else if (res.status === 404) {
          throw new Error("Drug not found"); 
        } else if (res.status === 500) {
          throw new Error("Something went wrong"); 
        }
        return res.json();
      })
      .then((data) => {
        setManualDrugData(data.data);
        setManualDrugPage(data.page);
        setManualDrugTotalPages(data.total_pages);
      })
      .catch((err) => {
        setManualDrugErrorMessage(err.message);
      })
    }
  }

  return (
    <div className={`flex-1 flex items-center justify-center`}>
      <div
        className={`
          grid grid-cols-1
          text-tn-d-fg
          border-3 pb-0 p-4
          w-full max-w-4xl
        `}
      >
        <h1 className={`text-2xl`}>Patient</h1>
        <div className={`border-1 mt-3 mb-3 pl-2 pr-2 pt-2 pb-2`}>
          <SearchBar 
              searchString={patientSearchString}
              setSearchString={setPatientSearchString}
              onSearch={onPatientSearch}
          />
          {
            errorMessage &&
              <div className={`text-base text-tn-d-red mt-1 mb-1`}>{errorMessage}</div>
          }
          <PatientDetails title="Username" value={patientData.username} />
          <PatientDetails title="Email" value={patientData.email} />
          <PatientDetails title="National ID" value={patientData.national_id} />
        </div>
        <h1 className={`text-2xl`}>Disease</h1>
        <div className={`border-1 mt-3 mb-3 pl-2 pr-2 pt-2 pb-2`}>
          <SearchBar 
              searchString={diseaseSearchString}
              setSearchString={setDiseaseSearchString}
              onSearch={onDiseaseSearch}
          />
          <DiseaseTable 
            diseases={diseaseData} 
            page={diseasePage} 
            totalPages={diseaseTotalPages} 
            errorMessage={diseaseErrorMessage}
            onPrev={() => {
              if (diseasePage > 1) {
                onDiseaseSearch(diseasePage - 1);
              }
            }}
            onNext={() => {
              if (diseasePage < diseaseTotalPages) {
                onDiseaseSearch(diseasePage + 1);
              }
            }}
            onSelect={handleDiseaseSelect}
          />
        </div>
        <h1 className={`text-2xl`}>Drugs</h1>
        <div className={`border-1 mt-3 mb-3 pl-2 pr-2 pt-2 pb-2`}>
          <h2 className={`text-m`}>Recommended</h2>
          <DrugTable 
            drugs={recDrugData} 
            page={recDrugPage}
            totalPages={recDrugTotalPages}
            errorMessage={recDrugErrorMessage}
            onPrev={ () => {
              if (!selectedDisease) return;
              recommendDrugs(selectedDisease, recDrugPage - 1, recOrderBy, recOrderMethod);
            } }
            onNext={ () => {
              if (!selectedDisease) return;
              recommendDrugs(selectedDisease, recDrugPage + 1, recOrderBy, recOrderMethod);
            } }
            onSelect={handleRecommendedDrugSelect}
            orderBy={recOrderBy}
            orderMethod={recOrderMethod}
            onSortChange={toggleRecSort}
          />
          <h2 className={`text-m`}>Manual search</h2>
          <SearchBar 
              searchString={drugSearchString}
              setSearchString={setDrugSearchString}
              onSearch={ () => {onDrugSearch(1, manualOrderBy, manualOrderMethod)} }
          />
          <DrugTable 
            drugs={manualDrugData}
            page={manualDrugPage}
            totalPages={manualDrugTotalPages}
            errorMessage={manualDrugErrorMessage}
            onPrev={ () => {onDrugSearch(manualDrugPage - 1, manualOrderBy, manualOrderMethod)} }
            onNext={ () => {onDrugSearch(manualDrugPage + 1, manualOrderBy, manualOrderMethod)} }
            onSelect={handleRecommendedDrugSelect}
            orderBy={manualOrderBy}
            orderMethod={manualOrderMethod}
            onSortChange={toggleManualSort}
          />
        </div>

      </div>
    </div> 
  )
}

function PatientDetails({ title, value }: { title: string, value: string}) {
  return (
    <div className={`grid grid-cols-2`}>
      <span className={`font-medium text-sm pt-2`}>{ title }:</span>
      <div className={`overflow-x-auto whitespace-nowrap`}>
        <span>{ value }</span>
      </div>
    </div> 
  );
}

export default Prescribe;
