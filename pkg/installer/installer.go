package installer

import (
	"github.com/pkg/errors"
	"github.com/pkosiec/terminer/internal/printer"
	"github.com/pkosiec/terminer/pkg/recipe"
	"github.com/pkosiec/terminer/pkg/shell"
)

// Installer provides an ability to install recipes
type Installer struct {
	r  *recipe.Recipe
	sh shell.Shell
}

// New creates a new instance of Installer.
func New(r *recipe.Recipe) (*Installer, error) {
	if r == nil {
		return nil, errors.New("Recipe is empty")
	}

	if err := r.Validate(); err != nil {
		return nil, err
	}

	return &Installer{
		r:  r,
		sh: shell.New(),
	}, nil
}

// Install installs a recipe by executing all steps in all stages
func (installer *Installer) Install() error {
	printer.RecipeInfo(installer.r, "Installing")

	stages := installer.r.Stages
	stagesLen := len(stages)

	for stageIndex, stage := range stages {
		printer.Stage(stage, stageIndex, stagesLen)

		stepsLen := len(stage.Steps)
		for stepIndex, step := range stage.Steps {
			printer.Step(step.Metadata, step.Execute.Run, stepIndex, stepsLen)

			res, err := installer.sh.Exec(step.Execute)
			printer.StepOutput(res)
			if err != nil {
				return errors.Wrapf(err, "while executing command from Stage '%s', Step '%s'", stage.Metadata.Name, step.Metadata.Name)
			}
		}
	}

	return nil
}

// Rollback reverts a recipe by executing all steps in all stages in reverse order
func (installer *Installer) Rollback() error {
	printer.RecipeInfo(installer.r, "Uninstalling")

	stages := installer.r.Stages
	stagesLen := len(stages)

	for i := stagesLen; i > 0; i-- {
		stage := stages[i-1]
		stageIndex := stagesLen - i
		printer.Stage(stage, stageIndex, stagesLen)

		stepsLen := len(stage.Steps)
		for j := stepsLen; j > 0; j-- {
			step := stage.Steps[j-1]
			stepIndex := stepsLen - j

			printer.Step(step.Metadata, step.Rollback.Run, stepIndex, stepsLen)

			res, err := installer.sh.Exec(step.Rollback)
			printer.StepOutput(res)
			if err != nil {
				// Print error and continue
				wrappedErr := errors.Wrapf(err, "while executing command from Stage %s, Step %s", stage.Metadata.Name, step.Metadata.Name)
				printer.Error(wrappedErr)
			}
		}
	}

	return nil
}
