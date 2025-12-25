import NoteTable from "./_components/noteTable"

export default function Home() {
  return (
    <>
      <table id="header">
        <tbody>
          <tr>
            <td className="logo">ntkpr</td>
          </tr>
          <tr>
            <td className="sub">TUI based note management system</td>
            <td></td>
          </tr>
        </tbody>
      </table>

      <table className="tabs">
        <tbody>
          <tr>
            <td>
              <a className="active" href="#">
                index
              </a>
            </td>
          </tr>
        </tbody>
      </table>

      <NoteTable />
      
    </>
  );
}
