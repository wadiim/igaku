import { useState } from 'react';

interface Drug {
  id: string,
  name: string,
  substance: string,
}

interface DrugTableProps {
  drugs: Drug[];
  page: number;
  totalPages: number;
  errorMessage: string | null;
  onPrev: () => void;
  onNext: () => void;
  onSelect: (selected: Drug[]) => void;
}

function DrugTable(
  { drugs, page, totalPages, errorMessage, onPrev, onNext, onSelect }: DrugTableProps 
) {
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())

  const handleSelect = (drug: Drug) => {
    const newSet = new Set(selectedIds);
    if (newSet.has(drug.id)) {
        newSet.delete(drug.id);
    } else {
        newSet.add(drug.id);
    }
    setSelectedIds(newSet);
    const selectedDrugs = drugs.filter(d => newSet.has(d.id));
    onSelect(selectedDrugs);
  };

  const listDrugs = drugs.map((drug, idx) =>
    <tr 
      key={drug.id}
      className={`bg-neutral-primary border-b border-default`}>
      <th 
        scope="row" 
        className={`
          px-6 py-4 
          font-medium text-heading whitespace-nowrap
          truncate
        `}
      >
        {drug.name}
      </th>
      <td className={`px-6 py-4 truncate`}>
        {drug.substance}
      </td>
      <td className={`px-6 py-4 text-center`}>
        <input
          type="checkbox"
          name="drug-select"
          checked={selectedIds.has(drug.id)}
          onChange={() => handleSelect(drug)}
          className={`
            border
            checked:border-brand
            focus:ring-brand-subtle 
          `}
        />
      </td>
    </tr>
  );

  return (
    <div className={`drug-pagination-table`}>
      <div 
        className={`
          relative 
          overflow-x-auto mb-2
          border
        `}
      >
        <table 
          className={`
            w-full text-sm text-left 
            rtl:text-right text-body
          `}
        >
          <thead className={`text-sm text-body border-b`}>
            <tr>
              <th scope="col" className={`px-6 py-3 font-medium w-2/5 truncate`}>
                Name
              </th>
              <th scope="col" className={`px-6 py-3 font-medium w-2/5 truncate`}>
                Substance
              </th>
              <th scope="col" className={`px-6 py-3 font-medium w-1/5 text-center`}>
                Select
              </th>
            </tr>
          </thead>
          <tbody>
          {drugs.length === 0 ? (
            <tr>
              <td colSpan={4} className="px-6 py-4 text-center text-gray-500">
                No drugs found. Try a search.
              </td>
            </tr>
          ) : (
            listDrugs
          )}
          </tbody>
        </table>
      </div>
      {
        errorMessage &&
          <div className={`text-base text-tn-d-red mt-1 mb-1`}>{errorMessage}</div>
      }

      <div className={`flex flex-col items-center`}>
        <span className={`text-sm text-body`}>
            Showing <span>{page}</span> of <span>{totalPages}</span> pages
        </span>
        <div className={`inline-flex mt-4`}>
          <button 
            type="button" 
            onClick={onPrev}
            disabled={page <= 1}
            className={`
              inline-flex items-center border 
              font-medium text-sm px-4 py-2.5
            `}
          >
            Previous
          </button>
          <button 
            type="button" 
            onClick={onNext}
            disabled={page >= totalPages}
            className={`
              inline-flex items-center border 
              font-medium text-sm px-4 py-2.5
            `}
          >
            Next
          </button>
        </div>
      </div>
    </div>
  );
}

export default DrugTable;
