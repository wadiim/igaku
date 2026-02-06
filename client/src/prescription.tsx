import { jwtDecode } from 'jwt-decode'
import { isTokenExpired } from './utils/auth';
import { useState, useEffect } from 'react';

interface Drug {
  id: string;
  rxcui: string;
  name: string;
  substance: string;
}

interface Prescription {
  id: string;
  PatientID: string;
  DoctorID: string;
  CreatedAt: string;
  Patient: {
    id: string;
    national_id: string;
  };
  Drugs: Drug[];
}

function PrescriptionsView() {
  const [prescriptions, setPrescriptions] = useState<Prescription[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchPrescriptions();
  }, []);
  

  const fetchPrescriptions = () => {
    let jwt = localStorage.getItem("jwt"); 
    if (isTokenExpired(jwt)) {
      navigate("/");
    }

    if (jwt === null) {
      throw new Error("Authentication failed");
    } else {
      let patientID = jwtDecode(jwt).sub
      fetch(`http://localhost:4000/med/history/${patientID}`, {
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
        } else if (res.status === 404) {
          throw new Error("Patient data not found"); 
        } else if (res.status === 500) {
          throw new Error("Something went wrong"); 
        }
        return res.json();
      })
      .then((data) => {
        setPrescriptions(data);
        localStorage.setItem("prescriptionData", JSON.stringify(data));
        setLoading(false);
      })
      .catch((err) => {
        console.log(err) 
        if (err instanceof TypeError && err.message === "Failed to fetch") {
          // No network connection.
          // NOTE: This detection mechanism does not work in Firefox.
          const catchedPrescriptionData = localStorage.getItem("prescriptionData");
          if (catchedPrescriptionData) {
            setPrescriptions(JSON.parse(catchedPrescriptionData));
            // setError(null);
          } else {
            // setError("Failed to load user data");
          }
          setLoading(false);
        } else {
          setError(err.message);
          setLoading(false);
        }
      });
    }
  }

  if (loading) return <div className="text-tn-d-fg p-8">Loading history...</div>;

  return (
    <div className="min-h-screen p-4 md:p-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold text-tn-d-fg mb-8">Medical History</h1>
        
        <div className="space-y-4">
          {prescriptions.length === 0 ? (
            <p className="text-gray-400">No prescriptions found.</p>
          ) : (
            prescriptions.map((item) => (
              <PrescriptionCard key={item.id} prescription={item} />
            ))
          )}
        </div>
      </div>
    </div>
  );
}

function PrescriptionCard({ prescription }: { prescription: Prescription }) {
  const date = new Date(prescription.CreatedAt).toLocaleDateString('en-GB', {
    day: '2-digit',
    month: 'long',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });

  return (
    <div className="border border-white overflow-hidden">
      <div className="p-4 flex justify-between items-center border-b border-white">
        <div>
          <h3 className="text-tn-d-fg">{date}</h3>
        </div>
        <div className="text-right">
          <p className="text-gray-500 text-xs uppercase">Prescription ID</p>
          <p className="text-gray-400 text-xs">{prescription.id}</p>
        </div>
      </div>

      <div className="p-4">
        <div className="mb-4">
          <span className="text-tn-d-fg text-sm uppercase">National ID: </span>
          <span className="text-tn-d-fg font-medium">{prescription.Patient.national_id}</span>
        </div>

        <h4 className="text-tn-d-fg text-sm font-bold mb-3 uppercase tracking-tighter">
          Prescribed Drugs ({prescription.Drugs.length})
        </h4>

        {prescription.Drugs.length > 0 ? (
          <div className="grid gap-3">
            {prescription.Drugs.map((drug) => (
              <div key={drug.id} className="bg-black/30 p-3 rounded border border-gray-800 flex flex-col sm:flex-row sm:justify-between">
                <div>
                  <p className="text-tn-d-fg font-medium">{drug.name}</p>
                  <p className="text-gray-500 text-xs italic">Substance: {drug.substance}</p>
                </div>
                <div className="mt-2 sm:mt-0 text-left sm:text-right">
                  <span className="bg-tn-d-dblue/20 text-tn-d-blue text-[10px] px-2 py-1 border border-tn-d-dblue/30">
                    RXCUI: {drug.rxcui}
                  </span>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-gray-600 italic text-sm">No specific medications listed for this entry.</p>
        )}
      </div>
    </div>
  );
}

export default PrescriptionsView;
